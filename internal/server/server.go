package server

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni/v3"

	"github.com/chainflag/eth-faucet/internal/chain"
	"github.com/chainflag/eth-faucet/web"
)

type Server struct {
	chain.TxBuilder
	cfg    *Config
	server *http.Server
	store  *Store // shared with the guard (windows, ETH and token payout logs)
}

func NewServer(builder chain.TxBuilder, cfg *Config) *Server {
	return &Server{
		TxBuilder: builder,
		cfg:       cfg,
	}
}

func (s *Server) setupRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", withNotFound(web.Dist()))
	p := s.cfg.protection
	ipr, err := NewIPResolver(p.TrustedProxies, p.CloudflareCIDRs)
	if err != nil {
		log.Fatalf("bad proxy config: %v", err)
	}
	ethCap := func(v float64) *big.Int {
		if v <= 0 {
			return nil
		}
		return chain.EtherToWei(v)
	}
	gid := p.GullyID
	if !p.GullyIDRequired {
		gid = nil
	}
	s.store = NewStore(p.StateFile)
	var devGID GullyIDVerifier
	if s.cfg.devTier.enabled() {
		devGID = p.GullyID
		if devGID == nil {
			devGID = NewGullyIDVerifier(s.cfg.gullyIDIssuer, s.cfg.gullyIDClient)
		}
	}
	guard := NewGuard(s.store, ipr, GuardOptions{
		Window: time.Duration(s.cfg.interval) * time.Minute, Payout: chain.EtherToWei(s.cfg.payout), Symbol: s.cfg.symbol,
		Deny: p.Deny, GullyID: gid, Activity: p.Activity, HourCap: ethCap(p.HourCapEth), DayCap: ethCap(p.DayCapEth), Alert: p.Alert,
		DevTier: s.cfg.devTier, DevGID: devGID,
	})
	hcaptcha := NewCaptcha(s.cfg.hcaptchaSiteKey, s.cfg.hcaptchaSecret)
	// captcha first: a bot without a captcha token never reaches the Gully ID / RPC checks
	router.Handle("/api/claim", negroni.New(hcaptcha, guard, negroni.Wrap(s.handleClaim())))
	router.Handle("/api/info", s.handleInfo())

	return router
}

func (s *Server) Run() {
	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(s.setupRouter())

	s.server = &http.Server{
		Addr:         ":" + strconv.Itoa(s.cfg.httpPort),
		Handler:      n,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Infof("Starting http server on port %d", s.cfg.httpPort)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

func (s *Server) handleClaim() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Address has already been validated by limiter
		address, _ := readAddress(r)

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		tier := tierFrom(r.Context(), s.cfg.payout)
		txHash, err := s.Transfer(ctx, address, tier.amount)
		if err != nil {
			log.WithFields(log.Fields{
				"error":   err,
				"address": address,
			}).Error("Failed to send transaction")
			renderJSON(w, claimResponse{Message: fmt.Sprintf("Transaction failed: %v", err)}, http.StatusInternalServerError)
			return
		}

		log.WithFields(log.Fields{
			"txHash":  txHash,
			"address": address,
		}).Info("Transaction sent successfully")
		resp := claimResponse{Message: fmt.Sprintf("Txhash: %s", txHash), Tier: tier.name, Amount: weiToEth(tier.amount)}
		if s.cfg.token != nil {
			tctx, tcancel := context.WithTimeout(r.Context(), 6*time.Second)
			resp.Token = s.dripToken(tctx, address)
			tcancel()
		}
		renderJSON(w, resp, http.StatusOK)
	}
}

func (c *Config) devTierInfo() *devTierInfo {
	if !c.devTier.enabled() {
		return nil
	}
	return &devTierInfo{Payout: strconv.FormatFloat(c.devTier.Payout, 'f', -1, 64), WindowHours: c.devTier.Window.Hours()}
}

func (s *Server) handleInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.NotFound(w, r)
			return
		}
		renderJSON(w, infoResponse{
			Account:         s.Sender().String(),
			Network:         s.cfg.network,
			Symbol:          s.cfg.symbol,
			Payout:          strconv.FormatFloat(s.cfg.payout, 'f', -1, 64),
			HcaptchaSiteKey: s.cfg.hcaptchaSiteKey,
			GullyIDClient:   s.cfg.gullyIDClient,
			GullyIDIssuer:   s.cfg.gullyIDIssuer,
			SignInRequired:  s.cfg.protection.GullyIDRequired,
			Eligibility:     s.cfg.protection.Activity != nil,
			WindowHours:     float64(s.cfg.interval) / 60,
			Token:           s.cfg.tokenInfo(),
			DevTier:         s.cfg.devTierInfo(),
		}, http.StatusOK)
	}
}
