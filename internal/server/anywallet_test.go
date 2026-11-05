package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/negroni/v3"
)

// Any wallet (owner 2026-11-05: "i should be able to send test eth to any wallet not just gullyid it needs to be
// option"). Sign-in with Gully ID is optional on the live faucet:
//   - no session: any address with history is paid (address + IP windows, denylist, budgets, eligibility check);
//   - a valid session: one claim per account per window too, and ANY recipient, a brand-new address included;
//   - a missing, expired or invalid session, or a Gully ID outage, never fails the claim: it is an anonymous claim.
// FAUCET_REQUIRE_GULLY_ID=true and the developer tier behave as before.
//
// This file only uses names that exist before the change, so on the old guard it compiles and fails on behaviour
// (the red check of the deploy script runs exactly that).

const (
	addrE = "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	addrF = "0x2546BcD3c84621e976D8185a91A922aE77ECEc30"

	freshOptionalMsg = "Sign in with Gully ID to fund a brand-new wallet, or use a wallet that has made at least one transaction on Gully Chain or Arbitrum Sepolia."
)

// downGID is a Gully ID that does not answer.
type downGID struct {
	mu    sync.Mutex
	calls int
}

func (d *downGID) Verify(context.Context, string) (*GullyUser, error) {
	d.mu.Lock()
	d.calls++
	d.mu.Unlock()
	return nil, errors.New("gully id unreachable: dial tcp: i/o timeout")
}

// newModeServer: 0.0005 ETH / 24 h. client "" = Gully ID not set up at all.
func newModeServer(t *testing.T, client string, p Protection, d *DevTier) (http.Handler, *MockTxBuilder, *Server) {
	t.Helper()
	m := new(MockTxBuilder)
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{1}, nil)
	m.On("Sender").Return(common.HexToAddress(addrA))
	cfg := &Config{httpPort: 8080, interval: 1440, network: "testnet", symbol: "ETH", payout: 0.0005, gullyIDClient: client}
	if client != "" {
		cfg.gullyIDIssuer = DefaultGullyIDIssuer
	}
	if err := cfg.SetProtection(p); err != nil {
		t.Fatal(err)
	}
	if d != nil {
		if err := cfg.SetDevTier(d); err != nil {
			t.Fatal(err)
		}
	}
	s := NewServer(m, cfg)
	n := negroni.New()
	n.UseHandler(s.setupRouter())
	return n, m, s
}

func transfers(m *MockTxBuilder) int {
	n := 0
	for _, c := range m.Calls {
		if c.Method == "Transfer" {
			n++
		}
	}
	return n
}

func hasLimit(s *Server, key string) bool {
	s.store.mu.Lock()
	defer s.store.mu.Unlock()
	_, ok := s.store.Limits[key]
	return ok
}

func modeInfo(t *testing.T, h http.Handler) map[string]interface{} {
	t.Helper()
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, httptest.NewRequest("GET", "/api/info", nil))
	var out map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func active(addrs ...string) *fakeActivity {
	a := &fakeActivity{active: map[string]bool{}}
	for _, x := range addrs {
		a.active[strings.ToLower(x)] = true
	}
	return a
}

func TestAnyWalletAnonymousAddressWithHistoryPays(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"tokA": {Sub: "user-a"}}}
	h, m, s := newModeServer(t, "gully-faucet", Protection{GullyID: gid, Activity: active(addrB)}, nil)
	if code, msg := claim(h, addrB, "", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatalf("anonymous, address with history: %d %s", code, msg)
	}
	if transfers(m) != 1 {
		t.Fatalf("transfers: %d", transfers(m))
	}
	// the anonymous windows still apply: same address, and same client IP with another address
	if code, _ := claim(h, addrB, "", "203.0.113.2", ""); code != http.StatusTooManyRequests {
		t.Fatalf("same address again: %d", code)
	}
	if code, _ := claim(h, addrC, "", "203.0.113.1", ""); code != http.StatusTooManyRequests {
		t.Fatalf("same IP again: %d", code)
	}
	if hasLimit(s, "gid:user-a") {
		t.Fatal("an anonymous claim must not touch an account window")
	}
}

func TestAnyWalletAnonymousFreshAddressRefusedWithSignInHint(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"tokA": {Sub: "user-a"}}}
	h, m, s := newModeServer(t, "gully-faucet", Protection{GullyID: gid, Activity: active(addrB)}, nil)
	code, msg := claim(h, addrA, "", "203.0.113.1", "")
	if code != http.StatusForbidden {
		t.Fatalf("anonymous, fresh address: %d %s", code, msg)
	}
	if !strings.Contains(msg, "no history") || !strings.Contains(msg, freshOptionalMsg) {
		t.Fatalf("refusal must say how to fund a brand-new wallet, got: %s", msg)
	}
	if strings.Contains(msg, "smart account of your Gully ID") {
		t.Fatalf("the required-mode wording leaked into optional mode: %s", msg)
	}
	if transfers(m) != 0 {
		t.Fatal("a refused claim sent ETH")
	}
	// a refusal burns no window: the same client then claims to an address with history
	if hasLimit(s, "addr:"+strings.ToLower(addrA)) {
		t.Fatal("refused claim left an address window")
	}
	if code, msg := claim(h, addrB, "", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatalf("same client, address with history: %d %s", code, msg)
	}
}

func TestAnyWalletOptionalSessionPaysFreshAddressAndUsesAccountWindow(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"tokA": {Sub: "user-a"}, "tokB": {Sub: "user-b"}}}
	h, m, s := newModeServer(t, "gully-faucet", Protection{GullyID: gid, Activity: active(addrB)}, nil)

	// a brand-new address that is not a wallet of this Gully ID: paid, because the account is verified
	if code, msg := claim(h, addrA, "tokA", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatalf("signed in, fresh address: %d %s", code, msg)
	}
	if !hasLimit(s, "gid:user-a") {
		t.Fatal("the claim did not consume the account window")
	}
	// second claim, same account: refused by the account window, even to another address from another IP
	code, msg := claim(h, addrB, "tokA", "203.0.113.2", "")
	if code != http.StatusTooManyRequests || !strings.Contains(msg, "Gully ID account already claimed") {
		t.Fatalf("second claim, same account: %d %s", code, msg)
	}
	if code, msg := claim(h, addrC, "tokA", "203.0.113.3", ""); code != http.StatusTooManyRequests || !strings.Contains(msg, "Gully ID account") {
		t.Fatalf("second claim, same account, fresh address: %d %s", code, msg)
	}
	// another account, another IP, another fresh address: paid
	if code, msg := claim(h, addrD, "tokB", "203.0.113.4", ""); code != http.StatusOK {
		t.Fatalf("other account: %d %s", code, msg)
	}
	// the address and IP windows hold for signed-in claims too
	gid.users["tokC"] = &GullyUser{Sub: "user-c"}
	if code, _ := claim(h, addrA, "tokC", "203.0.113.5", ""); code != http.StatusTooManyRequests {
		t.Fatalf("address window for a signed-in claim: %d", code)
	}
	if code, _ := claim(h, addrE, "tokC", "203.0.113.1", ""); code != http.StatusTooManyRequests {
		t.Fatalf("IP window for a signed-in claim: %d", code)
	}
	if transfers(m) != 2 {
		t.Fatalf("transfers: %d, want 2", transfers(m))
	}
}

func TestAnyWalletInvalidOrExpiredBearerIsAnAnonymousClaim(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"tokA": {Sub: "user-a"}}}
	h, _, s := newModeServer(t, "gully-faucet", Protection{GullyID: gid, Activity: active(addrB)}, nil)

	// fakeGID answers errNotSignedIn for a token it does not know, as Gully ID does for an expired or revoked one
	if code, msg := claim(h, addrB, "expired-token", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatalf("expired session, address with history must be paid as anonymous: %d %s", code, msg)
	}
	for k := range s.store.Limits {
		if strings.HasPrefix(k, "gid:") {
			t.Fatalf("an invalid session created an account window: %s", k)
		}
	}
	// and it unlocks nothing: a fresh address is refused exactly like an anonymous claim
	code, msg := claim(h, addrA, "expired-token", "203.0.113.2", "")
	if code != http.StatusForbidden || !strings.Contains(msg, freshOptionalMsg) {
		t.Fatalf("expired session, fresh address: %d %s", code, msg)
	}
	// never 401: sign-in is optional
	if code, _ := claim(h, addrC, strings.Repeat("x", 5000), "203.0.113.3", ""); code == http.StatusUnauthorized {
		t.Fatal("an oversized bearer answered 401 in optional mode")
	}
}

func TestAnyWalletGullyIDOutageDoesNotBlockAnonymousClaims(t *testing.T) {
	down := &downGID{}
	h, m, _ := newModeServer(t, "gully-faucet", Protection{GullyID: down, Activity: active(addrB, addrC)}, nil)

	// no bearer: Gully ID is not even asked
	if code, msg := claim(h, addrB, "", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatalf("anonymous claim during a Gully ID outage: %d %s", code, msg)
	}
	if down.calls != 0 {
		t.Fatalf("Gully ID was called %d times for a claim without a session", down.calls)
	}
	// a bearer that can not be checked: the claim goes on as anonymous and is paid (the address has history)
	if code, msg := claim(h, addrC, "tokA", "203.0.113.2", ""); code != http.StatusOK {
		t.Fatalf("session during an outage, address with history: %d %s", code, msg)
	}
	// a fresh address needs the sign-in, which can not be checked now: refused, saying so, and never a 503
	code, msg := claim(h, addrA, "tokA", "203.0.113.3", "")
	if code != http.StatusForbidden || !strings.Contains(msg, "no history") || !strings.Contains(msg, "could not check your Gully ID sign-in") {
		t.Fatalf("session during an outage, fresh address: %d %s", code, msg)
	}
	if transfers(m) != 2 {
		t.Fatalf("transfers: %d, want 2", transfers(m))
	}
}

func TestAnyWalletRequiredModeUnchanged(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"t1": {Sub: "u1"}, "t2": {Sub: "u2"}, "t3": {Sub: "u3"}}}
	gid.users["t3"].Wallets = append(gid.users["t3"].Wallets, struct {
		ChainID *int64 `json:"chainId"`
		Address string `json:"address"`
		Kind    string `json:"kind"`
	}{Address: addrC, Kind: "kernel-v3.3"})
	h, _, _ := newModeServer(t, "gully-faucet", Protection{GullyIDRequired: true, GullyID: gid, Activity: active(addrB)}, nil)

	if code, msg := claim(h, addrB, "", "203.0.113.1", ""); code != http.StatusUnauthorized || !strings.Contains(msg, "Sign in with Gully") {
		t.Fatalf("required, no session: %d %s", code, msg)
	}
	if code, _ := claim(h, addrB, "bogus", "203.0.113.1", ""); code != http.StatusUnauthorized {
		t.Fatalf("required, bad session: %d", code)
	}
	// a verified user still can not pay a fresh address that is not their own smart account
	code, msg := claim(h, addrA, "t1", "203.0.113.1", "")
	if code != http.StatusForbidden || !strings.Contains(msg, "smart account of your Gully ID") {
		t.Fatalf("required, fresh address: %d %s", code, msg)
	}
	if strings.Contains(msg, "brand-new wallet") {
		t.Fatalf("the optional wording leaked into required mode: %s", msg)
	}
	if code, _ := claim(h, addrB, "t2", "203.0.113.2", ""); code != http.StatusOK {
		t.Fatalf("required, address with history: %d", code)
	}
	if code, _ := claim(h, addrC, "t3", "203.0.113.3", ""); code != http.StatusOK {
		t.Fatalf("required, own smart account: %d", code)
	}
	if code, msg := claim(h, addrD, "t2", "203.0.113.4", ""); code != http.StatusTooManyRequests || !strings.Contains(msg, "Gully ID account") {
		t.Fatalf("required, account window: %d %s", code, msg)
	}
	// required + Gully ID down: no claim (503), as before
	h2, m2, _ := newModeServer(t, "gully-faucet", Protection{GullyIDRequired: true, GullyID: &downGID{}, Activity: active(addrB)}, nil)
	if code, _ := claim(h2, addrB, "t1", "203.0.113.1", ""); code != http.StatusServiceUnavailable {
		t.Fatalf("required, Gully ID down: %d", code)
	}
	if transfers(m2) != 0 {
		t.Fatal("required mode paid during a Gully ID outage")
	}
}

func TestAnyWalletDeveloperTierUnchanged(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"dev": {Sub: "dev-1"}, "user": {Sub: "user-1"}}}
	for _, required := range []bool{true, false} {
		h, m, s := newModeServer(t, "gully-faucet", Protection{GullyIDRequired: required, GullyID: gid, Activity: active(addrB)}, devTier("dev-1"))
		// a developer: the larger drip to a fresh address, on the tier's own keys (no IP key)
		code, r := devClaim(h, addrA, "dev", "203.0.113.1")
		if code != http.StatusOK || r.Tier != "developer" || r.Amount != "0.05" {
			t.Fatalf("required=%v developer claim: %d %+v", required, code, r)
		}
		if !hasLimit(s, "dev:gid:dev-1") || !hasLimit(s, "dev:addr:"+strings.ToLower(addrA)) || hasLimit(s, "gid:dev-1") {
			t.Fatalf("required=%v developer keys: %v", required, s.store.Limits)
		}
		for k := range s.store.Limits {
			if strings.HasPrefix(k, "ip:") {
				t.Fatalf("required=%v developer claim took an IP window: %s", required, k)
			}
		}
		// a normal user never gets the developer amount
		code, r = devClaim(h, addrB, "user", "203.0.113.2")
		if code != http.StatusOK || r.Tier != "standard" || r.Amount != "0.0005" {
			t.Fatalf("required=%v normal user: %d %+v", required, code, r)
		}
		if got := sentAmounts(m); len(got) != 2 {
			t.Fatalf("required=%v amounts %v", required, got)
		}
	}
	// optional mode + the tier on: no session = standard tier, and a non-developer session gets the standard tier with
	// the account window, on any address
	h, _, s := newModeServer(t, "gully-faucet", Protection{GullyID: gid, Activity: active(addrB)}, devTier("dev-1"))
	if code, r := devClaim(h, addrB, "", "203.0.113.1"); code != http.StatusOK || r.Tier != "standard" {
		t.Fatalf("optional + tier, anonymous: %d %+v", code, r)
	}
	if code, r := devClaim(h, addrC, "user", "203.0.113.2"); code != http.StatusOK || r.Tier != "standard" || !hasLimit(s, "gid:user-1") {
		t.Fatalf("optional + tier, normal user to a fresh address: %d %+v", code, r)
	}
}

func TestAnyWalletDenylistAndBudgetsApplyInEveryMode(t *testing.T) {
	type mode struct {
		name     string
		client   string
		required bool
		token    string // "" = anonymous
	}
	modes := []mode{
		{"off", "", false, ""},
		{"optional anonymous", "gully-faucet", false, ""},
		{"optional signed in", "gully-faucet", false, "tok"},
		{"optional bad session", "gully-faucet", false, "bogus"},
		{"required", "gully-faucet", true, "tok"},
	}
	for _, md := range modes {
		// one account per claim, so only the denylist and the budget can refuse
		gid := &fakeGID{users: map[string]*GullyUser{}}
		n := 0
		tok := func() string {
			if md.token != "tok" {
				return md.token
			}
			n++
			name := "tok-" + string(rune('a'+n))
			gid.users[name] = &GullyUser{Sub: "user-" + name}
			return name
		}
		al := &countAlert{}
		p := Protection{GullyIDRequired: md.required, Deny: map[string]bool{strings.ToLower(addrD): true}, HourCapEth: 0.001, DayCapEth: 0.01,
			Activity: active(addrA, addrB, addrC, addrD, addrE, addrF), Alert: al}
		if md.client != "" {
			p.GullyID = gid
		}
		h, m, _ := newModeServer(t, md.client, p, nil)

		code, msg := claim(h, addrD, tok(), "203.0.113.1", "")
		if code != http.StatusForbidden || !strings.Contains(msg, "not eligible") {
			t.Fatalf("%s: denylisted address: %d %s", md.name, code, msg)
		}
		if code, msg := claim(h, addrA, tok(), "203.0.113.2", ""); code != http.StatusOK {
			t.Fatalf("%s: claim 1: %d %s", md.name, code, msg)
		}
		if code, msg := claim(h, addrB, tok(), "203.0.113.3", ""); code != http.StatusOK {
			t.Fatalf("%s: claim 2: %d %s", md.name, code, msg)
		}
		// hourly budget = two payouts
		code, msg = claim(h, addrC, tok(), "203.0.113.4", "")
		if code != http.StatusServiceUnavailable || !strings.Contains(msg, "hourly budget") {
			t.Fatalf("%s: over the hourly budget: %d %s", md.name, code, msg)
		}
		if transfers(m) != 2 {
			t.Fatalf("%s: transfers %d, want 2", md.name, transfers(m))
		}
		if al.n != 1 {
			t.Fatalf("%s: budget alerts %d, want 1", md.name, al.n)
		}
	}

	// the daily budget, signed in to fresh addresses in optional mode (the path with the fewest other checks)
	gid := &fakeGID{users: map[string]*GullyUser{"a": {Sub: "a"}, "b": {Sub: "b"}, "c": {Sub: "c"}}}
	h, _, _ := newModeServer(t, "gully-faucet", Protection{GullyID: gid, DayCapEth: 0.001, Activity: active()}, nil)
	for i, tk := range []string{"a", "b"} {
		if code, msg := claim(h, []string{addrA, addrB}[i], tk, "203.0.113."+string(rune('1'+i)), ""); code != http.StatusOK {
			t.Fatalf("daily budget, claim %d: %d %s", i, code, msg)
		}
	}
	if code, msg := claim(h, addrC, "c", "203.0.113.9", ""); code != http.StatusServiceUnavailable || !strings.Contains(msg, "daily budget") {
		t.Fatalf("over the daily budget: %d %s", code, msg)
	}
}

func TestAnyWalletFailedSendFreesTheAccountWindow(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"tokA": {Sub: "user-a"}}}
	m := new(MockTxBuilder)
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{}, errors.New("boom")).Once()
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{1}, nil)
	m.On("Sender").Return(common.HexToAddress(addrA))
	cfg := &Config{httpPort: 8080, interval: 1440, network: "testnet", symbol: "ETH", payout: 0.0005, gullyIDClient: "gully-faucet", gullyIDIssuer: DefaultGullyIDIssuer}
	if err := cfg.SetProtection(Protection{GullyID: gid, Activity: active()}); err != nil {
		t.Fatal(err)
	}
	s := NewServer(m, cfg)
	n := negroni.New()
	n.UseHandler(s.setupRouter())
	if code, _ := claim(n, addrA, "tokA", "203.0.113.1", ""); code != http.StatusInternalServerError {
		t.Fatalf("failed send: %d", code)
	}
	if hasLimit(s, "gid:user-a") {
		t.Fatal("a failed send kept the account window")
	}
	if code, msg := claim(n, addrA, "tokA", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatalf("retry after a failed send: %d %s", code, msg)
	}
	if !hasLimit(s, "gid:user-a") {
		t.Fatal("the paid retry did not take the account window")
	}
}

func TestAnyWalletInfoTellsThePageTheMode(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{}}
	cases := []struct {
		name, client string
		required     bool
		mode         string
	}{
		{"required", "gully-faucet", true, "required"},
		{"optional", "gully-faucet", false, "optional"},
		{"off", "", false, "off"},
	}
	for _, c := range cases {
		p := Protection{GullyIDRequired: c.required, Activity: active()}
		if c.client != "" {
			p.GullyID = gid
		}
		h, _, _ := newModeServer(t, c.client, p, nil)
		info := modeInfo(t, h)
		if info["signin_mode"] != c.mode {
			t.Fatalf("%s: signin_mode = %v, want %s", c.name, info["signin_mode"], c.mode)
		}
		if info["signin_required"] != c.required {
			t.Fatalf("%s: signin_required = %v", c.name, info["signin_required"])
		}
		if info["eligibility_check"] != true {
			t.Fatalf("%s: eligibility_check = %v", c.name, info["eligibility_check"])
		}
	}
}

// No Gully ID at all (mode "off") and the eligibility check on: the refusal must not tell people to sign in.
func TestAnyWalletOffModeFreshAddressMessage(t *testing.T) {
	h, _, _ := newModeServer(t, "", Protection{Activity: active(addrB)}, nil)
	code, msg := claim(h, addrA, "whatever", "203.0.113.1", "")
	if code != http.StatusForbidden || !strings.Contains(msg, "no history") || strings.Contains(msg, "Sign in") || strings.Contains(msg, "Gully ID") {
		t.Fatalf("off mode, fresh address: %d %s", code, msg)
	}
}
