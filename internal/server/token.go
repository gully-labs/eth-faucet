package server

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"

	"github.com/chainflag/eth-faucet/internal/chain"
)

// TokenDrip is the optional test-stablecoin payout (glUSD on Gully Chain) that goes out with every successful ETH
// claim. It sits behind the same gate as the ETH payout (sign-in, eligibility, denylist, real IP, per-account/address/IP
// window, ETH budget), and has its own budget on top: a rolling 24 h cap and an optional lifetime cap. It fails
// closed and never takes the ETH claim down with it: an empty token budget, a short wallet or a failed send only adds
// a note to the claim result.
type TokenDrip struct {
	Token      common.Address
	Symbol     string
	Decimals   uint8
	Amount     *big.Int // base units per claim
	DayCap     *big.Int // base units per rolling 24 h; required
	TotalCap   *big.Int // base units over the faucet's lifetime; nil = no lifetime cap
	AmountText string   // "1"
}

// NewTokenDrip builds the drip from human amounts ("1" glUSD). dayCap must be > 0; totalCap 0 = no lifetime cap.
func NewTokenDrip(token, symbol string, decimals uint8, amount, dayCap, totalCap float64) (*TokenDrip, error) {
	if !common.IsHexAddress(token) {
		return nil, fmt.Errorf("token drip: %q is not an address", token)
	}
	if amount <= 0 {
		return nil, errors.New("token drip: amount must be > 0")
	}
	if dayCap < amount {
		return nil, errors.New("token drip: a daily token budget (>= one payout) is required")
	}
	if totalCap != 0 && totalCap < amount {
		return nil, errors.New("token drip: the lifetime token budget must be 0 (off) or >= one payout")
	}
	units := func(v float64) *big.Int {
		d := decimal.NewFromFloat(v).Shift(int32(decimals)).Truncate(0)
		return d.BigInt()
	}
	t := &TokenDrip{Token: common.HexToAddress(token), Symbol: symbol, Decimals: decimals, Amount: units(amount),
		DayCap: units(dayCap), AmountText: strconv.FormatFloat(amount, 'f', -1, 64)}
	if totalCap > 0 {
		t.TotalCap = units(totalCap)
	}
	if t.Amount.Sign() <= 0 {
		return nil, errors.New("token drip: amount is below one base unit")
	}
	return t, nil
}

func (t *TokenDrip) human(v *big.Int) string {
	return decimal.NewFromBigInt(v, -int32(t.Decimals)).String()
}

// tokenResult is the token part of a successful claim response.
type tokenResult struct {
	Symbol string `json:"symbol"`
	Amount string `json:"amount"`
	Sent   bool   `json:"sent"`
	Tx     string `json:"tx,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

var tokenSeq int64

// reserveTokenLocked checks both token budgets and books one payout. "" ref = refused, with the reason.
func (s *Server) reserveTokenLocked(t *TokenDrip) (string, string) {
	st := s.store
	if spent := st.tokenSpentLocked(24 * time.Hour); new(big.Int).Add(spent, t.Amount).Cmp(t.DayCap) > 0 {
		return "", fmt.Sprintf("Today's %s budget is used up, so this claim is ETH only. Try again tomorrow.", t.Symbol)
	}
	if t.TotalCap != nil && new(big.Int).Add(st.tokenTotalLocked(), t.Amount).Cmp(t.TotalCap) > 0 {
		return "", fmt.Sprintf("The faucet's %s allowance is used up, so this claim is ETH only.", t.Symbol)
	}
	ref := "t" + strconv.FormatInt(atomic.AddInt64(&tokenSeq, 1), 10) + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	st.TokenPayouts = append(st.TokenPayouts, payoutRecord{At: st.now().Unix(), Wei: t.Amount.String(), Ref: ref})
	st.TokenTotal = new(big.Int).Add(st.tokenTotalLocked(), t.Amount).String()
	st.saveLocked()
	return ref, ""
}

func (s *Server) releaseToken(t *TokenDrip, ref string) {
	st := s.store
	st.mu.Lock()
	defer st.mu.Unlock()
	for i, p := range st.TokenPayouts {
		if p.Ref == ref {
			st.TokenPayouts = append(st.TokenPayouts[:i], st.TokenPayouts[i+1:]...)
			tot := new(big.Int).Sub(st.tokenTotalLocked(), t.Amount)
			if tot.Sign() < 0 {
				tot.SetInt64(0)
			}
			st.TokenTotal = tot.String()
			break
		}
	}
	st.saveLocked()
}

// dripToken runs after the ETH payout was sent. It never fails the claim.
func (s *Server) dripToken(ctx context.Context, address string) *tokenResult {
	t := s.cfg.token
	ts, ok := s.TxBuilder.(chain.TokenSender)
	if t == nil || !ok || s.store == nil {
		return nil
	}
	res := &tokenResult{Symbol: t.Symbol, Amount: t.AmountText}
	logf := log.WithFields(log.Fields{"address": address, "token": t.Symbol})

	s.store.mu.Lock()
	ref, refusal := s.reserveTokenLocked(t)
	s.store.mu.Unlock()
	if ref == "" {
		logf.Warn("token drip skipped: budget")
		res.Msg = refusal
		return res
	}
	bal, err := ts.TokenBalance(ctx, t.Token, s.Sender())
	if err != nil || bal.Cmp(t.Amount) < 0 {
		s.releaseToken(t, ref)
		if err != nil {
			logf.WithError(err).Warn("token drip skipped: balance check failed")
		} else {
			logf.WithField("balance", t.human(bal)).Warn("token drip skipped: faucet wallet is out of the token")
			if s.cfg.protection.Alert != nil {
				s.cfg.protection.Alert.Alert(fmt.Sprintf("Gully faucet: the hot wallet has %s %s left, below one payout of %s. Token drip paused; ETH claims continue.", t.human(bal), t.Symbol, t.AmountText))
			}
		}
		res.Msg = fmt.Sprintf("The faucet's %s is empty right now, so this claim is ETH only.", t.Symbol)
		return res
	}
	hash, err := ts.TransferToken(ctx, t.Token, address, t.Amount)
	if err != nil {
		s.releaseToken(t, ref)
		logf.WithError(err).Error("token drip failed")
		res.Msg = fmt.Sprintf("The %s transfer failed, so this claim is ETH only.", t.Symbol)
		return res
	}
	logf.WithField("txHash", hash.Hex()).Info("token drip sent")
	res.Sent, res.Tx = true, hash.Hex()
	return res
}

// tokenInfo is the token part of /api/info.
type tokenInfo struct {
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals uint8  `json:"decimals"`
	Amount   string `json:"amount"`
}

func (c *Config) tokenInfo() *tokenInfo {
	if c.token == nil {
		return nil
	}
	return &tokenInfo{Address: c.token.Token.Hex(), Symbol: c.token.Symbol, Decimals: c.token.Decimals, Amount: c.token.AmountText}
}

// SetTokenDrip turns the token drip on (nil = off).
func (c *Config) SetTokenDrip(t *TokenDrip) {
	if t != nil && strings.TrimSpace(t.Symbol) == "" {
		t.Symbol = "TOKEN"
	}
	c.token = t
}
