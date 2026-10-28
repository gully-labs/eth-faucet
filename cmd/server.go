package cmd

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"

	"github.com/chainflag/eth-faucet/internal/chain"
	"github.com/chainflag/eth-faucet/internal/server"
)

var (
	appVersion = "v1.2.1"
	chainIDMap = map[string]int{"sepolia": 11155111, "holesky": 17000}

	httpPortFlag = flag.Int("httpport", 8080, "Listener port to serve HTTP connection")
	proxyCntFlag = flag.Int("proxycount", 0, "Count of reverse proxies in front of the server")
	versionFlag  = flag.Bool("version", false, "Print version number")

	payoutFlag   = flag.Float64("faucet.amount", 1, "Number of Ethers to transfer per user request")
	intervalFlag = flag.Int("faucet.minutes", 1440, "Number of minutes to wait between funding rounds")
	netnameFlag  = flag.String("faucet.name", "testnet", "Network name to display on the frontend")
	symbolFlag   = flag.String("faucet.symbol", "ETH", "Token symbol to display on the frontend")

	keyJSONFlag  = flag.String("wallet.keyjson", os.Getenv("KEYSTORE"), "Keystore file to fund user requests with")
	keyPassFlag  = flag.String("wallet.keypass", "password.txt", "Passphrase text file to decrypt keystore")
	privKeyFlag  = flag.String("wallet.privkey", os.Getenv("PRIVATE_KEY"), "Private key hex to fund user requests with")
	providerFlag = flag.String("wallet.provider", os.Getenv("WEB3_PROVIDER"), "Ethereum JSON-RPC endpoint, or a comma-separated ordered list with automatic fallback (empty = the Gully Chain testnet public list)")

	hcaptchaSiteKeyFlag = flag.String("hcaptcha.sitekey", os.Getenv("HCAPTCHA_SITEKEY"), "hCaptcha sitekey")
	hcaptchaSecretFlag  = flag.String("hcaptcha.secret", os.Getenv("HCAPTCHA_SECRET"), "hCaptcha secret")

	gullyIDClientFlag = flag.String("gullyid.client", os.Getenv("GULLY_ID_CLIENT_ID"), "Gully ID client id for the \"Sign in with Gully\" button (empty = off)")
	gullyIDIssuerFlag = flag.String("gullyid.issuer", os.Getenv("GULLY_ID_ISSUER"), "Gully ID issuer origin (default "+server.DefaultGullyIDIssuer+")")

	// drain hardening (2026-10-27)
	gullyIDRequiredFlag = flag.Bool("gullyid.required", envBool("FAUCET_REQUIRE_GULLY_ID", false), "Claims need a Sign in with Gully session (one claim per Gully ID account per window)")
	eligibilityFlag     = flag.Bool("faucet.eligibility", envBool("FAUCET_ELIGIBILITY", false), "Refuse fresh recipients: the address must be the user's Gully smart account or have history on one of -faucet.activityrpcs")
	activityRPCsFlag    = flag.String("faucet.activityrpcs", os.Getenv("FAUCET_ACTIVITY_RPCS"), "Comma-separated RPCs for the eligibility check (empty = the faucet RPC + "+defaultArbSepoliaRPC+")")
	hourCapFlag         = flag.Float64("faucet.hourcap", envFloat("FAUCET_BUDGET_HOUR", 0), "Max ETH paid out per rolling hour (0 = no cap)")
	dayCapFlag          = flag.Float64("faucet.daycap", envFloat("FAUCET_BUDGET_DAY", 0), "Max ETH paid out per rolling 24 h (0 = no cap)")
	stateFileFlag       = flag.String("faucet.statefile", os.Getenv("FAUCET_STATE_FILE"), "JSON file that keeps rate limits and payouts across restarts (empty = memory only)")
	denylistFlag        = flag.String("faucet.denylist", os.Getenv("FAUCET_DENYLIST_FILE"), "File with denied recipient addresses, one per line")
	trustedProxiesFlag  = flag.String("proxy.trusted", os.Getenv("FAUCET_TRUSTED_PROXIES"), "CIDRs of reverse proxies whose X-Forwarded-For is read (empty = loopback + private ranges)")
	cloudflareFlag      = flag.String("proxy.cloudflare", os.Getenv("FAUCET_CLOUDFLARE_CIDRS"), "CIDRs whose CF-Connecting-IP is trusted (empty = Cloudflare's published ranges)")
	alertWebhookFlag    = flag.String("alert.webhook", os.Getenv("FAUCET_ALERT_WEBHOOK"), "URL that receives budget alerts as JSON {text}")
	tgChatFlag          = flag.String("alert.telegramchat", os.Getenv("TELEGRAM_CHAT_ID"), "Telegram chat id for alerts (bot token: TELEGRAM_BOT_TOKEN)")

	// token drip (2026-10-28): a small test-stablecoin payout (glUSD) with every successful ETH claim
	tokenFlag         = flag.String("token.address", os.Getenv("FAUCET_TOKEN_ADDRESS"), "ERC-20 sent with every successful claim (empty = ETH only)")
	tokenSymbolFlag   = flag.String("token.symbol", envString("FAUCET_TOKEN_SYMBOL", "glUSD"), "Symbol of -token.address shown on the page")
	tokenAmountFlag   = flag.Float64("token.amount", envFloat("FAUCET_TOKEN_AMOUNT", 1), "Tokens per claim (whole units)")
	tokenDayCapFlag   = flag.Float64("token.daycap", envFloat("FAUCET_TOKEN_BUDGET_DAY", 0), "Max tokens per rolling 24 h (required with -token.address)")
	tokenTotalCapFlag = flag.Float64("token.totalcap", envFloat("FAUCET_TOKEN_BUDGET_TOTAL", 0), "Max tokens over the faucet's lifetime, kept in the state file (0 = no lifetime cap)")
)

const defaultArbSepoliaRPC = "https://sepolia-rollup.arbitrum.io/rpc"

func envBool(k string, def bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(k)))
	switch v {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	}
	return def
}

func envString(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func envFloat(k string, def float64) float64 {
	if v, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv(k)), 64); err == nil {
		return v
	}
	return def
}

func init() {
	flag.Parse()
	if *versionFlag {
		fmt.Println(appVersion)
		os.Exit(0)
	}
}

func Execute() {
	privateKey, err := getPrivateKeyFromFlags()
	if err != nil {
		panic(fmt.Errorf("failed to read private key: %w", err))
	}
	var chainID *big.Int
	if value, ok := chainIDMap[strings.ToLower(*netnameFlag)]; ok {
		chainID = big.NewInt(int64(value))
	}

	txBuilder, err := chain.NewTxBuilder(*providerFlag, privateKey, chainID)
	if err != nil {
		panic(fmt.Errorf("cannot connect to web3 provider: %w", err))
	}

	config := server.NewConfig(*netnameFlag, *symbolFlag, *httpPortFlag, *intervalFlag, *proxyCntFlag, *payoutFlag, *hcaptchaSiteKeyFlag, *hcaptchaSecretFlag)
	if err := config.SetGullyID(*gullyIDClientFlag, *gullyIDIssuerFlag); err != nil {
		panic(err)
	}
	deny, err := server.LoadDenylist(*denylistFlag)
	if err != nil {
		panic(fmt.Errorf("denylist: %w", err))
	}
	prot := server.Protection{
		GullyIDRequired: *gullyIDRequiredFlag,
		HourCapEth:      *hourCapFlag,
		DayCapEth:       *dayCapFlag,
		StateFile:       *stateFileFlag,
		Deny:            deny,
		TrustedProxies:  *trustedProxiesFlag,
		CloudflareCIDRs: *cloudflareFlag,
		Alert:           server.NewAlerter(os.Getenv("TELEGRAM_BOT_TOKEN"), *tgChatFlag, *alertWebhookFlag),
	}
	if *eligibilityFlag {
		urls := chain.ParseProviders(*activityRPCsFlag)
		if strings.TrimSpace(*activityRPCsFlag) == "" {
			urls = append(chain.ParseProviders(*providerFlag), defaultArbSepoliaRPC)
		}
		prot.Activity = server.NewActivityChecker(urls)
	}
	if err := config.SetProtection(prot); err != nil {
		panic(err)
	}
	log.WithFields(log.Fields{
		"signin_required": prot.GullyIDRequired, "eligibility": prot.Activity != nil, "hour_cap": prot.HourCapEth,
		"day_cap": prot.DayCapEth, "state_file": prot.StateFile, "denylist": len(deny),
		"telegram_alerts": os.Getenv("TELEGRAM_BOT_TOKEN") != "" && *tgChatFlag != "",
	}).Info("faucet protection")
	if strings.TrimSpace(*tokenFlag) != "" {
		drip, err := newTokenDrip(txBuilder)
		if err != nil {
			panic(err)
		}
		config.SetTokenDrip(drip)
		log.WithFields(log.Fields{"token": drip.Token.Hex(), "symbol": drip.Symbol, "amount": drip.AmountText,
			"day_cap": *tokenDayCapFlag, "total_cap": *tokenTotalCapFlag}).Info("token drip on")
	}
	srv := server.NewServer(txBuilder, config)

	// Run server in goroutine
	go srv.Run()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	// Graceful shutdown
	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	} else {
		log.Info("Server exited")
	}
}

func getPrivateKeyFromFlags() (*ecdsa.PrivateKey, error) {
	if *privKeyFlag != "" {
		hexkey := *privKeyFlag
		if chain.Has0xPrefix(hexkey) {
			hexkey = hexkey[2:]
		}
		return crypto.HexToECDSA(hexkey)
	} else if *keyJSONFlag == "" {
		return nil, errors.New("missing private key or keystore")
	}

	keyfile, err := chain.ResolveKeyfilePath(*keyJSONFlag)
	if err != nil {
		return nil, err
	}
	password, err := os.ReadFile(*keyPassFlag)
	if err != nil {
		return nil, err
	}

	return chain.DecryptKeyfile(keyfile, strings.TrimRight(string(password), "\r\n"))
}

func newTokenDrip(b chain.TxBuilder) (*server.TokenDrip, error) {
	ts, ok := b.(chain.TokenSender)
	if !ok {
		return nil, errors.New("token drip: this wallet backend cannot send tokens")
	}
	if !common.IsHexAddress(*tokenFlag) {
		return nil, fmt.Errorf("token drip: %q is not an address", *tokenFlag)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	dec, err := ts.TokenDecimals(ctx, common.HexToAddress(*tokenFlag))
	if err != nil {
		return nil, fmt.Errorf("token drip: read decimals: %w", err)
	}
	return server.NewTokenDrip(*tokenFlag, *tokenSymbolFlag, dec, *tokenAmountFlag, *tokenDayCapFlag, *tokenTotalCapFlag)
}
