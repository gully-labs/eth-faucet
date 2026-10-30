package server

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/negroni/v3"
)

const usdce = "0x65Ba03c6FA90Ed549f465D4c8f2e784D6eEB7E76"

// tokenBuilder is a TxBuilder that also sends tokens.
type tokenBuilder struct {
	*MockTxBuilder
	mu      sync.Mutex
	bal     *big.Int
	balErr  error
	sendErr error
	sent    map[string]*big.Int
}

func (b *tokenBuilder) TokenBalance(context.Context, common.Address, common.Address) (*big.Int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.balErr != nil {
		return nil, b.balErr
	}
	return new(big.Int).Set(b.bal), nil
}

func (b *tokenBuilder) TokenDecimals(context.Context, common.Address) (uint8, error) { return 6, nil }

func (b *tokenBuilder) TransferToken(_ context.Context, _ common.Address, to string, amount *big.Int) (common.Hash, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.sendErr != nil {
		return common.Hash{}, b.sendErr
	}
	b.bal.Sub(b.bal, amount)
	b.sent[strings.ToLower(to)] = amount
	return common.Hash{2}, nil
}

type fullResp struct {
	Msg   string       `json:"msg"`
	Token *tokenResult `json:"token"`
}

func newTokenServer(t *testing.T, p Protection, drip *TokenDrip, bal int64, ethOK bool) (http.Handler, *tokenBuilder) {
	t.Helper()
	m := new(MockTxBuilder)
	var err error
	if !ethOK {
		err = errors.New("boom")
	}
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{1}, err)
	m.On("Sender").Return(common.HexToAddress(addrA))
	b := &tokenBuilder{MockTxBuilder: m, bal: big.NewInt(bal), sent: map[string]*big.Int{}}
	cfg := &Config{httpPort: 8080, interval: 1440, network: "testnet", symbol: "ETH", payout: 0.0005, gullyIDClient: "gully-faucet"}
	if err := cfg.SetProtection(p); err != nil {
		t.Fatal(err)
	}
	cfg.SetTokenDrip(drip)
	s := NewServer(b, cfg)
	n := negroni.New()
	n.UseHandler(s.setupRouter())
	return n, b
}

func claimFull(h http.Handler, addr, token, cfIP string) (int, fullResp) {
	req := httptest.NewRequest("POST", "/api/claim", strings.NewReader(`{"address":"`+addr+`"}`))
	req.RemoteAddr = "172.17.0.2:5555"
	req.Header.Set("X-Forwarded-For", "162.158.94.11")
	req.Header.Set("CF-Connecting-IP", cfIP)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	var r fullResp
	_ = json.Unmarshal(rr.Body.Bytes(), &r)
	return rr.Code, r
}

func mustDrip(t *testing.T, day, total float64) *TokenDrip {
	t.Helper()
	d, err := NewTokenDrip(usdce, "USDC.e", 6, 1, day, total)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestNewTokenDripValidation(t *testing.T) {
	if _, err := NewTokenDrip(usdce, "USDC.e", 6, 1, 0, 0); err == nil {
		t.Fatal("a token drip without a daily budget must be refused")
	}
	if _, err := NewTokenDrip("nope", "USDC.e", 6, 1, 5, 0); err == nil {
		t.Fatal("bad address")
	}
	if _, err := NewTokenDrip(usdce, "USDC.e", 6, 1, 5, 0.5); err == nil {
		t.Fatal("lifetime cap below one payout")
	}
	d := mustDrip(t, 3, 5)
	if d.Amount.String() != "1000000" || d.DayCap.String() != "3000000" || d.TotalCap.String() != "5000000" {
		t.Fatalf("units: %s %s %s", d.Amount, d.DayCap, d.TotalCap)
	}
}

func TestTokenDripWithClaim(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"tokA": {Sub: "user-a"}, "tokB": {Sub: "user-b"}}}
	h, b := newTokenServer(t, Protection{GullyIDRequired: true, GullyID: gid}, mustDrip(t, 5, 0), 10_000_000, true)

	// not signed in: no ETH and no token
	if code, _ := claimFull(h, addrA, "", "203.0.113.1"); code != http.StatusUnauthorized || len(b.sent) != 0 {
		t.Fatalf("unsigned: %d sent=%d", code, len(b.sent))
	}
	code, r := claimFull(h, addrB, "tokA", "203.0.113.1")
	if code != http.StatusOK || r.Token == nil || !r.Token.Sent || r.Token.Symbol != "USDC.e" || r.Token.Amount != "1" || r.Token.Tx == "" {
		t.Fatalf("claim: %d %+v", code, r.Token)
	}
	if v := b.sent[strings.ToLower(addrB)]; v == nil || v.String() != "1000000" {
		t.Fatalf("sent %v", v)
	}
	// same account again: refused by the shared window, no second token
	if code, _ := claimFull(h, addrC, "tokA", "203.0.113.2"); code != http.StatusTooManyRequests || len(b.sent) != 1 {
		t.Fatalf("second: %d sent=%d", code, len(b.sent))
	}
}

func TestTokenDripFailsClosedEthStillPays(t *testing.T) {
	// daily token budget = 2 payouts
	h, b := newTokenServer(t, Protection{}, mustDrip(t, 2, 0), 10_000_000, true)
	for i, a := range []string{addrA, addrB} {
		if code, r := claimFull(h, a, "", "203.0.113."+string(rune('1'+i))); code != http.StatusOK || !r.Token.Sent {
			t.Fatalf("%d: %d %+v", i, code, r.Token)
		}
	}
	code, r := claimFull(h, addrC, "", "203.0.113.9")
	if code != http.StatusOK || !strings.HasPrefix(r.Msg, "Txhash:") || r.Token == nil || r.Token.Sent || !strings.Contains(r.Token.Msg, "budget") {
		t.Fatalf("over token budget must still pay ETH: %d %q %+v", code, r.Msg, r.Token)
	}
	if len(b.sent) != 2 {
		t.Fatalf("sent %d", len(b.sent))
	}
}

func TestTokenDripLifetimeCapAndPersistence(t *testing.T) {
	state := filepath.Join(t.TempDir(), "state.json")
	h, _ := newTokenServer(t, Protection{StateFile: state}, mustDrip(t, 10, 1), 10_000_000, true)
	if _, r := claimFull(h, addrA, "", "203.0.113.1"); !r.Token.Sent {
		t.Fatal("first")
	}
	// restart: the lifetime total survives
	h2, b2 := newTokenServer(t, Protection{StateFile: state}, mustDrip(t, 10, 1), 10_000_000, true)
	code, r := claimFull(h2, addrB, "", "203.0.113.2")
	if code != http.StatusOK || r.Token.Sent || !strings.Contains(r.Token.Msg, "allowance") || len(b2.sent) != 0 {
		t.Fatalf("lifetime cap after restart: %d %+v", code, r.Token)
	}
}

func TestTokenDripEmptyWalletAndErrors(t *testing.T) {
	h, b := newTokenServer(t, Protection{}, mustDrip(t, 10, 0), 500_000, true) // half a USDC.e left
	code, r := claimFull(h, addrA, "", "203.0.113.1")
	if code != http.StatusOK || r.Token.Sent || !strings.Contains(r.Token.Msg, "empty") {
		t.Fatalf("empty wallet: %d %+v", code, r.Token)
	}
	b.bal.SetInt64(10_000_000)
	b.balErr = errors.New("rpc down")
	if code, r := claimFull(h, addrB, "", "203.0.113.2"); code != http.StatusOK || r.Token.Sent {
		t.Fatalf("rpc down: %d %+v", code, r.Token)
	}
	b.balErr = nil
	b.sendErr = errors.New("nonce too low")
	if code, r := claimFull(h, addrC, "", "203.0.113.3"); code != http.StatusOK || r.Token.Sent || !strings.Contains(r.Token.Msg, "failed") {
		t.Fatalf("send error: %d %+v", code, r.Token)
	}
	// refused / failed drips do not use up the token budget: 10 a day, nothing booked yet
	b.sendErr = nil
	if code, r := claimFull(h, addrD, "", "203.0.113.4"); code != http.StatusOK || !r.Token.Sent {
		t.Fatalf("after recovery: %d %+v", code, r.Token)
	}
}

func TestNoTokenWhenEthFails(t *testing.T) {
	h, b := newTokenServer(t, Protection{}, mustDrip(t, 10, 0), 10_000_000, false)
	if code, r := claimFull(h, addrA, "", "203.0.113.1"); code != http.StatusInternalServerError || r.Token != nil || len(b.sent) != 0 {
		t.Fatalf("eth failed: %d %+v", code, r.Token)
	}
}

func TestInfoShowsToken(t *testing.T) {
	h, _ := newTokenServer(t, Protection{}, mustDrip(t, 10, 0), 0, true)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/api/info", nil))
	var info infoResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &info)
	if info.Token == nil || info.Token.Symbol != "USDC.e" || info.Token.Amount != "1" || info.Token.Decimals != 6 {
		t.Fatalf("info: %s", rr.Body.String())
	}
}
