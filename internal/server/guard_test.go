package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/negroni/v3"

	"github.com/chainflag/eth-faucet/internal/chain"
)

type fakeGID struct{ users map[string]*GullyUser }

func (f *fakeGID) Verify(_ context.Context, tok string) (*GullyUser, error) {
	if u, ok := f.users[tok]; ok {
		return u, nil
	}
	return nil, errNotSignedIn
}

type fakeActivity struct {
	active map[string]bool
	err    error
}

func (f *fakeActivity) HasActivity(_ context.Context, a string) (bool, error) {
	return f.active[strings.ToLower(a)], f.err
}

type countAlert struct {
	mu sync.Mutex
	n  int
}

func (c *countAlert) Alert(string) { c.mu.Lock(); c.n++; c.mu.Unlock() }

const (
	addrA = "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"
	addrB = "0x8F5461E7F422F0737f7BcA44ED23226F3c9e2e50"
	addrC = "0xf9C889D547a264ed5d2239070bc9bd1940E34f2C"
	addrD = "0x3CCD2234a6A11e9501936f9be54dffd476E2194d"
)

func newTestGuardServer(t *testing.T, p Protection, okSend bool) (http.Handler, *MockTxBuilder) {
	t.Helper()
	m := new(MockTxBuilder)
	var err error
	if !okSend {
		err = errors.New("boom")
	}
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{1}, err)
	m.On("Sender").Return(common.HexToAddress(addrA))
	cfg := &Config{httpPort: 8080, interval: 1440, network: "testnet", symbol: "ETH", payout: 0.0005, gullyIDClient: "gully-faucet"}
	if err := cfg.SetProtection(p); err != nil {
		t.Fatal(err)
	}
	s := NewServer(m, cfg)
	n := negroni.New()
	n.UseHandler(s.setupRouter())
	return n, m
}

func claim(h http.Handler, addr, token, cfIP, peer string) (int, string) {
	req := httptest.NewRequest("POST", "/api/claim", strings.NewReader(`{"address":"`+addr+`"}`))
	req.RemoteAddr = "172.17.0.2:5555" // the edge Caddy container
	if peer == "" {
		peer = "162.158.94.11" // a Cloudflare edge
	}
	req.Header.Set("X-Forwarded-For", peer)
	if cfIP != "" {
		req.Header.Set("CF-Connecting-IP", cfIP)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	var r claimResponse
	_ = json.Unmarshal(rr.Body.Bytes(), &r)
	return rr.Code, r.Message
}

func TestClientIP(t *testing.T) {
	r, _ := NewIPResolver("", "")
	mk := func(remote, xff, cf string) *http.Request {
		q := httptest.NewRequest("GET", "/", nil)
		q.RemoteAddr = remote
		if xff != "" {
			q.Header.Set("X-Forwarded-For", xff)
		}
		if cf != "" {
			q.Header.Set("CF-Connecting-IP", cf)
		}
		return q
	}
	cases := []struct{ remote, xff, cf, want string }{
		{"172.17.0.2:1", "162.158.94.11", "203.0.113.7", "203.0.113.7"},          // Cloudflare -> Caddy -> faucet
		{"172.17.0.2:1", "1.2.3.4, 162.158.94.11", "203.0.113.7", "203.0.113.7"}, // client-supplied XFF ignored
		{"172.17.0.2:1", "198.51.100.9", "203.0.113.7", "198.51.100.9"},          // origin hit directly: CF header forged
		{"198.51.100.9:1", "162.158.94.11", "203.0.113.7", "198.51.100.9"},       // untrusted peer: headers ignored
		{"172.17.0.2:1", "", "", "172.17.0.2"},
	}
	for _, c := range cases {
		if got := r.ClientIP(mk(c.remote, c.xff, c.cf)); got != c.want {
			t.Errorf("%v: got %s", c, got)
		}
	}
	if b := ipBucket("2001:db8:1:2:3:4:5:6"); b != "2001:db8:1:2::/64" {
		t.Errorf("v6 bucket %s", b)
	}
}

func TestSignInRequiredAndPerAccountLimit(t *testing.T) {
	gid := &fakeGID{users: map[string]*GullyUser{"tokA": {Sub: "user-a"}}}
	h, _ := newTestGuardServer(t, Protection{GullyIDRequired: true, GullyID: gid}, true)

	if code, msg := claim(h, addrA, "", "203.0.113.1", ""); code != http.StatusUnauthorized || !strings.Contains(msg, "Sign in with Gully") {
		t.Fatalf("no token: %d %s", code, msg)
	}
	if code, _ := claim(h, addrA, "bogus", "203.0.113.1", ""); code != http.StatusUnauthorized {
		t.Fatalf("bad token: %d", code)
	}
	if code, msg := claim(h, addrA, "tokA", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatalf("first claim: %d %s", code, msg)
	}
	// same account, other address, other IP: refused by the per-account window
	if code, msg := claim(h, addrB, "tokA", "203.0.113.2", ""); code != http.StatusTooManyRequests || !strings.Contains(msg, "Gully ID") {
		t.Fatalf("second claim same account: %d %s", code, msg)
	}
}

func TestRealIPLimitAcrossCloudflareEdges(t *testing.T) {
	h, _ := newTestGuardServer(t, Protection{}, true)
	if code, _ := claim(h, addrA, "", "203.0.113.9", "162.158.94.11"); code != http.StatusOK {
		t.Fatal("first")
	}
	// the drain pattern: new address, different Cloudflare edge, same real client
	if code, _ := claim(h, addrB, "", "203.0.113.9", "104.23.199.157"); code != http.StatusTooManyRequests {
		t.Fatalf("same client IP via another edge must be limited, got %d", code)
	}
}

func TestEligibility(t *testing.T) {
	smart := strings.ToLower(addrC)
	gid := &fakeGID{users: map[string]*GullyUser{
		"t1": {Sub: "u1"}, "t2": {Sub: "u2"}, "t3": {Sub: "u3"},
	}}
	gid.users["t3"].Wallets = append(gid.users["t3"].Wallets, struct {
		ChainID *int64 `json:"chainId"`
		Address string `json:"address"`
		Kind    string `json:"kind"`
	}{Address: addrC, Kind: "kernel-v3.3"})
	act := &fakeActivity{active: map[string]bool{strings.ToLower(addrB): true}}
	h, _ := newTestGuardServer(t, Protection{GullyIDRequired: true, GullyID: gid, Activity: act}, true)

	if code, msg := claim(h, addrA, "t1", "203.0.113.1", ""); code != http.StatusForbidden || !strings.Contains(msg, "no history") {
		t.Fatalf("fresh EOA: %d %s", code, msg)
	}
	if code, _ := claim(h, addrB, "t2", "203.0.113.2", ""); code != http.StatusOK {
		t.Fatalf("active address: %d", code)
	}
	if code, _ := claim(h, addrC, "t3", "203.0.113.3", ""); code != http.StatusOK || smart == "" {
		t.Fatalf("own smart account: %d", code)
	}
	// RPC down: fail closed, and the failed attempt does not burn the window
	act.err = errors.New("rpc down")
	if code, _ := claim(h, addrD, "t1", "203.0.113.4", ""); code != http.StatusServiceUnavailable {
		t.Fatalf("rpc down: %d", code)
	}
}

func TestDenylistAndBudget(t *testing.T) {
	al := &countAlert{}
	deny := map[string]bool{strings.ToLower(addrD): true}
	// hourly cap = 2 payouts of 0.0005
	h, _ := newTestGuardServer(t, Protection{Deny: deny, HourCapEth: 0.001, Alert: al}, true)
	if code, _ := claim(h, addrD, "", "203.0.113.1", ""); code != http.StatusForbidden {
		t.Fatalf("denylisted: %d", code)
	}
	if code, _ := claim(h, addrA, "", "203.0.113.2", ""); code != http.StatusOK {
		t.Fatal("1")
	}
	if code, _ := claim(h, addrB, "", "203.0.113.3", ""); code != http.StatusOK {
		t.Fatal("2")
	}
	code, msg := claim(h, addrC, "", "203.0.113.4", "")
	if code != http.StatusServiceUnavailable || !strings.Contains(msg, "hourly budget") {
		t.Fatalf("over budget: %d %s", code, msg)
	}
	claim(h, addrC, "", "203.0.113.5", "")
	if al.n != 1 {
		t.Fatalf("expected exactly one alert, got %d", al.n)
	}
}

func TestFailedSendRollsBack(t *testing.T) {
	m := new(MockTxBuilder)
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{}, errors.New("boom")).Once()
	m.On("Transfer", mock.Anything, mock.Anything, mock.Anything).Return(common.Hash{1}, nil)
	cfg := &Config{httpPort: 8080, interval: 1440, network: "testnet", symbol: "ETH", payout: 0.0005}
	if err := cfg.SetProtection(Protection{HourCapEth: 0.0005}); err != nil {
		t.Fatal(err)
	}
	n := negroni.New()
	n.UseHandler(NewServer(m, cfg).setupRouter())
	if code, _ := claim(n, addrA, "", "203.0.113.1", ""); code != http.StatusInternalServerError {
		t.Fatalf("send fails: %d", code)
	}
	// window and budget were rolled back: the retry goes through
	if code, msg := claim(n, addrA, "", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatalf("retry: %d %s", code, msg)
	}
}

func TestStatePersistsAcrossRestart(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "state.json")
	h, _ := newTestGuardServer(t, Protection{StateFile: file}, true)
	if code, _ := claim(h, addrA, "", "203.0.113.1", ""); code != http.StatusOK {
		t.Fatal("first")
	}
	if _, err := os.Stat(file); err != nil {
		t.Fatal(err)
	}
	// "restart": a fresh server on the same file
	h2, _ := newTestGuardServer(t, Protection{StateFile: file}, true)
	if code, _ := claim(h2, addrA, "", "203.0.113.2", ""); code != http.StatusTooManyRequests {
		t.Fatalf("limit lost on restart: %d", code)
	}
	if code, _ := claim(h2, addrB, "", "203.0.113.1", ""); code != http.StatusTooManyRequests {
		t.Fatalf("ip limit lost on restart: %d", code)
	}
}

func TestStoreWindowExpiry(t *testing.T) {
	s := NewStore("")
	now := time.Unix(1_800_000_000, 0)
	s.now = func() time.Time { return now }
	s.Limits["k"] = now.Add(time.Minute).Unix()
	if s.remainingLocked("k") <= 0 {
		t.Fatal("should be limited")
	}
	now = now.Add(2 * time.Minute)
	if s.remainingLocked("k") != 0 {
		t.Fatal("should be free")
	}
	s.Payouts = []payoutRecord{{At: now.Add(-30 * time.Minute).Unix(), Wei: chain.EtherToWei(0.1).String()}, {At: now.Add(-2 * time.Hour).Unix(), Wei: chain.EtherToWei(0.2).String()}}
	if s.spentLocked(time.Hour).Cmp(chain.EtherToWei(0.1)) != 0 {
		t.Fatal("hour sum")
	}
}

func TestLoadDenylist(t *testing.T) {
	f := filepath.Join(t.TempDir(), "deny.txt")
	_ = os.WriteFile(f, []byte("# farmers\n"+addrA+"\n  "+strings.ToLower(addrB)+" # dust hop\nnot-an-address\n"), 0o600)
	d, err := LoadDenylist(f)
	if err != nil || len(d) != 2 || !d[strings.ToLower(addrA)] {
		t.Fatalf("%v %v", d, err)
	}
	if d, err := LoadDenylist(filepath.Join(t.TempDir(), "missing")); err != nil || len(d) != 0 {
		t.Fatal("missing file")
	}
}
