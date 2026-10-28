package server

import (
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/negroni/v3"

	"github.com/chainflag/eth-faucet/internal/chain"
)

// newDevTierServer: 0.0005 ETH / 24 h standard, developer tier 0.05 ETH / 60 min.
func newDevTierServer(t *testing.T, p Protection, d *DevTier) (http.Handler, *MockTxBuilder, *Server) {
	t.Helper()
	m := new(MockTxBuilder)
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{1}, nil)
	m.On("Sender").Return(common.HexToAddress(addrA))
	cfg := &Config{httpPort: 8080, interval: 1440, network: "testnet", symbol: "ETH", payout: 0.0005, gullyIDClient: "gully-faucet", gullyIDIssuer: DefaultGullyIDIssuer}
	if err := cfg.SetProtection(p); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetDevTier(d); err != nil {
		t.Fatal(err)
	}
	s := NewServer(m, cfg)
	n := negroni.New()
	n.UseHandler(s.setupRouter())
	return n, m, s
}

func devTier(subs ...string) *DevTier {
	d := &DevTier{Payout: 0.05, Window: time.Hour, Subs: map[string]bool{}, Wallets: map[string]bool{}}
	for _, s := range subs {
		d.Subs[s] = true
	}
	return d
}

func devClaim(h http.Handler, addr, token, cfIP string) (int, claimResponse) {
	req := httptest.NewRequest("POST", "/api/claim", strings.NewReader(`{"address":"`+addr+`"}`))
	req.RemoteAddr = "172.17.0.2:5555"
	req.Header.Set("X-Forwarded-For", "162.158.94.11")
	req.Header.Set("CF-Connecting-IP", cfIP)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	var r claimResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &r)
	return rr.Code, r
}

func sentAmounts(m *MockTxBuilder) []*big.Int {
	var out []*big.Int
	for _, c := range m.Calls {
		if c.Method == "Transfer" {
			out = append(out, c.Arguments.Get(2).(*big.Int))
		}
	}
	return out
}

func TestDevTierLargerDripSkipsFreshAddressCheck(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"dev": {Sub: "did:gully:dev"}, "user": {Sub: "did:gully:user"}}}
	act := &fakeActivity{active: map[string]bool{}} // every address is fresh
	h, m, _ := newDevTierServer(t, Protection{GullyIDRequired: true, GullyID: gid, Activity: act}, devTier("did:gully:dev"))

	code, r := devClaim(h, addrA, "dev", "203.0.113.1")
	if code != http.StatusOK || r.Tier != "developer" || r.Amount != "0.05" {
		t.Fatalf("developer claim: %d %+v", code, r)
	}
	if got := sentAmounts(m); len(got) != 1 || got[0].Cmp(chain.EtherToWei(0.05)) != 0 {
		t.Fatalf("sent %v, want 0.05 ETH", got)
	}
	// a normal user still gets the fresh-address refusal and never the developer amount
	if code, r := devClaim(h, addrB, "user", "203.0.113.2"); code != http.StatusForbidden || !strings.Contains(r.Message, "no history") {
		t.Fatalf("normal user, fresh address: %d %+v", code, r)
	}
	// hCaptcha + sign-in stay: no token is still refused
	if code, _ := devClaim(h, addrC, "", "203.0.113.3"); code != http.StatusUnauthorized {
		t.Fatalf("no sign-in: %d", code)
	}
}

func TestDevTierShorterWindowOwnKeysNoIPKey(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"d1": {Sub: "dev-1"}, "d2": {Sub: "dev-2"}, "u": {Sub: "user"}}}
	h, _, s := newDevTierServer(t, Protection{GullyIDRequired: true, GullyID: gid}, devTier("dev-1", "dev-2"))
	now := time.Now()
	s.store.now = func() time.Time { return now }

	if code, r := devClaim(h, addrA, "d1", "203.0.113.7"); code != http.StatusOK || r.Tier != "developer" {
		t.Fatalf("d1 first: %d %+v", code, r)
	}
	// same office IP: a second developer and a normal user can still claim (no IP key for the tier)
	if code, _ := devClaim(h, addrB, "d2", "203.0.113.7"); code != http.StatusOK {
		t.Fatalf("d2 same IP: %d", code)
	}
	if code, r := devClaim(h, addrC, "u", "203.0.113.7"); code != http.StatusOK || r.Tier != "standard" || r.Amount != "0.0005" {
		t.Fatalf("normal user same IP: %d %+v", code, r)
	}
	// d1 again inside the developer window: refused, with the tier's message
	if code, r := devClaim(h, addrD, "d1", "203.0.113.8"); code != http.StatusTooManyRequests || !strings.Contains(r.Message, "Developer tier") {
		t.Fatalf("d1 inside window: %d %+v", code, r)
	}
	// after the developer window (1 h, not 24 h): allowed again
	now = now.Add(61 * time.Minute)
	if code, _ := devClaim(h, addrD, "d1", "203.0.113.8"); code != http.StatusOK {
		t.Fatalf("d1 after window: %d", code)
	}
}

func TestDevTierBudgets(t *testing.T) {
	al := &countAlert{}
	gid := &fakeGID{users: map[string]*GullyUser{"d1": {Sub: "dev-1"}, "d2": {Sub: "dev-2"}, "d3": {Sub: "dev-3"}, "u": {Sub: "user"}}}
	d := devTier("dev-1", "dev-2", "dev-3")
	d.DayCap = 0.1 // two developer claims
	h, _, _ := newDevTierServer(t, Protection{GullyIDRequired: true, GullyID: gid, Alert: al}, d)
	for i, tok := range []string{"d1", "d2"} {
		if code, _ := devClaim(h, []string{addrA, addrB}[i], tok, "203.0.113.1"); code != http.StatusOK {
			t.Fatalf("dev claim %d: %d", i, code)
		}
	}
	if code, r := devClaim(h, addrC, "d3", "203.0.113.1"); code != http.StatusServiceUnavailable || !strings.Contains(r.Message, "developer tier") {
		t.Fatalf("over developer budget: %d %+v", code, r)
	}
	if al.n == 0 {
		t.Fatal("developer budget alert not sent")
	}
	// the standard tier is not blocked by the developer budget
	if code, _ := devClaim(h, addrD, "u", "203.0.113.2"); code != http.StatusOK {
		t.Fatalf("standard claim: %d", code)
	}

	// the global hourly cap also counts developer payouts
	gid2 := &fakeGID{users: map[string]*GullyUser{"d1": {Sub: "dev-1"}}}
	h2, _, _ := newDevTierServer(t, Protection{GullyIDRequired: true, GullyID: gid2, HourCapEth: 0.01}, devTier("dev-1"))
	if code, r := devClaim(h2, addrA, "d1", "203.0.113.1"); code != http.StatusServiceUnavailable || !strings.Contains(r.Message, "hourly budget") {
		t.Fatalf("developer claim over the global cap: %d %+v", code, r)
	}
}

func TestDevTierFailedSendRollsBack(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"d1": {Sub: "dev-1"}}}
	m := new(MockTxBuilder)
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{}, errors.New("boom")).Once()
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{1}, nil)
	m.On("Sender").Return(common.HexToAddress(addrA))
	cfg := &Config{httpPort: 8080, interval: 1440, network: "testnet", symbol: "ETH", payout: 0.0005, gullyIDClient: "gully-faucet"}
	if err := cfg.SetProtection(Protection{GullyIDRequired: true, GullyID: gid}); err != nil {
		t.Fatal(err)
	}
	if err := cfg.SetDevTier(devTier("dev-1")); err != nil {
		t.Fatal(err)
	}
	s := NewServer(m, cfg)
	n := negroni.New()
	n.UseHandler(s.setupRouter())
	if code, _ := devClaim(n, addrA, "d1", "203.0.113.1"); code != http.StatusInternalServerError {
		t.Fatalf("failed send: %d", code)
	}
	if code, r := devClaim(n, addrA, "d1", "203.0.113.1"); code != http.StatusOK || r.Tier != "developer" {
		t.Fatalf("retry after failed send: %d %+v", code, r)
	}
	if len(s.store.Payouts) != 1 || s.store.Payouts[0].Tier != "dev" {
		t.Fatalf("payout log: %+v", s.store.Payouts)
	}
}

func TestDevTierQualifiers(t *testing.T) {
	yes := true
	wallet := func(a string) struct {
		ChainID *int64 `json:"chainId"`
		Address string `json:"address"`
		Kind    string `json:"kind"`
	} {
		return struct {
			ChainID *int64 `json:"chainId"`
			Address string `json:"address"`
			Kind    string `json:"kind"`
		}{Address: a, Kind: "kernel-v3.3"}
	}
	d := devTier("sub-1")
	ParseDevList("# team\n0x"+strings.Repeat("aB", 20)+", sub-2\n", d.Subs, d.Wallets)
	withWallet := &GullyUser{Sub: "w"}
	withWallet.Wallets = append(withWallet.Wallets, wallet("0x"+strings.Repeat("Ab", 20)))
	cases := []struct {
		name        string
		u           *GullyUser
		addr        string
		trust, want bool
	}{
		{"nobody signed in", nil, addrA, true, false},
		{"allowlisted sub", &GullyUser{Sub: "sub-1"}, addrA, false, true},
		{"sub from the list string", &GullyUser{Sub: "sub-2"}, addrA, false, true},
		{"allowlisted Gully wallet", withWallet, addrA, false, true},
		{"allowlisted recipient", &GullyUser{Sub: "x"}, "0x" + strings.Repeat("ab", 20), false, true},
		{"profile developer, not trusted", &GullyUser{Sub: "p", Developer: &yes}, addrA, false, false},
		{"profile developer, trusted", &GullyUser{Sub: "p", Developer: &yes}, addrA, true, true},
		{"profile role, trusted", &GullyUser{Sub: "r", Roles: []string{"Developer"}}, addrA, true, true},
		{"plain user", &GullyUser{Sub: "z"}, addrA, true, false},
	}
	for _, c := range cases {
		d.TrustProfile = c.trust
		if got, _ := d.Match(c.u, c.addr); got != c.want {
			t.Errorf("%s: got %v", c.name, got)
		}
	}
	var off *DevTier
	if ok, _ := off.Match(&GullyUser{Sub: "sub-1"}, addrA); ok {
		t.Error("nil tier matched")
	}
}

func TestDevTierConfigAndInfo(t *testing.T) {
	cfg := &Config{interval: 1440, payout: 0.0005}
	if err := cfg.SetDevTier(devTier("a")); err == nil || !strings.Contains(err.Error(), "Gully ID") {
		t.Fatalf("needs Gully ID: %v", err)
	}
	cfg.gullyIDClient = "gully-faucet"
	if err := cfg.SetDevTier(&DevTier{Payout: 0.05, Window: time.Hour}); err == nil || !strings.Contains(err.Error(), "nobody can qualify") {
		t.Fatalf("empty allowlist: %v", err)
	}
	if err := cfg.SetDevTier(&DevTier{Payout: 0.05, Window: 48 * time.Hour, TrustProfile: true}); err == nil {
		t.Fatal("window longer than the normal window must be refused")
	}
	if err := cfg.SetDevTier(&DevTier{Payout: 0}); err != nil || cfg.devTier != nil || cfg.devTierInfo() != nil {
		t.Fatalf("amount 0 = off: %v", err)
	}

	gid := &fakeGID{users: map[string]*GullyUser{}}
	h, _, _ := newDevTierServer(t, Protection{GullyIDRequired: true, GullyID: gid}, devTier("a"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/api/info", nil))
	var info infoResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &info); err != nil {
		t.Fatal(err)
	}
	if info.DevTier == nil || info.DevTier.Payout != "0.05" || info.DevTier.WindowHours != 1 || info.Payout != "0.0005" {
		t.Fatalf("info: %s", rr.Body.String())
	}
}

func TestDevTierWithOptionalSignIn(t *testing.T) {
	// sign-in not required: a valid session upgrades to the developer tier, no session = standard tier
	gid := &fakeGID{users: map[string]*GullyUser{"d1": {Sub: "dev-1"}}}
	h, m, _ := newDevTierServer(t, Protection{GullyID: gid}, devTier("dev-1"))
	if code, r := devClaim(h, addrA, "", "203.0.113.1"); code != http.StatusOK || r.Tier != "standard" {
		t.Fatalf("anonymous: %d %+v", code, r)
	}
	if code, r := devClaim(h, addrB, "d1", "203.0.113.2"); code != http.StatusOK || r.Tier != "developer" {
		t.Fatalf("developer session: %d %+v", code, r)
	}
	if code, r := devClaim(h, addrC, "bogus", "203.0.113.3"); code != http.StatusOK || r.Tier != "standard" {
		t.Fatalf("bad session falls back to standard: %d %+v", code, r)
	}
	got := sentAmounts(m)
	if len(got) != 3 || got[1].Cmp(chain.EtherToWei(0.05)) != 0 || got[0].Cmp(chain.EtherToWei(0.0005)) != 0 {
		t.Fatalf("amounts %v", got)
	}
}
