package chain

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// rpcServer answers eth_chainId with chainID, unless fail says otherwise, and counts its hits.
func rpcServer(t *testing.T, hits *int32, fail func(w http.ResponseWriter) bool, chainID string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(hits, 1)
		if fail != nil && fail(w) {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"` + chainID + `"}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func chainIDVia(t *testing.T, ft *FallbackTransport) (uint64, error) {
	t.Helper()
	rc, err := rpc.DialHTTPWithClient(ft.URLs[0], &http.Client{Transport: ft})
	if err != nil {
		t.Fatal(err)
	}
	id, err := ethclient.NewClient(rc).ChainID(context.Background())
	if err != nil {
		return 0, err
	}
	return id.Uint64(), nil
}

func TestParseProviders(t *testing.T) {
	if got := ParseProviders(""); !reflect.DeepEqual(got, DefaultProviders) {
		t.Errorf("empty = %v, want defaults", got)
	}
	got := ParseProviders(" http://a , ,http://b,")
	if !reflect.DeepEqual(got, []string{"http://a", "http://b"}) {
		t.Errorf("list = %v", got)
	}
}

func TestFallbackOnErrors(t *testing.T) {
	cases := map[string]func(w http.ResponseWriter) bool{
		"http 503": func(w http.ResponseWriter) bool { w.WriteHeader(503); return true },
		"http 429": func(w http.ResponseWriter) bool { w.WriteHeader(429); return true },
		"json-rpc rate limit": func(w http.ResponseWriter) bool {
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32005,"message":"request rate limited"}}`))
			return true
		},
		"rate limit message": func(w http.ResponseWriter) bool {
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"Too Many Requests"}}`))
			return true
		},
	}
	for name, fail := range cases {
		t.Run(name, func(t *testing.T) {
			var badHits, goodHits int32
			bad := rpcServer(t, &badHits, fail, "0x1")
			good := rpcServer(t, &goodHits, nil, "0x768d6")
			id, err := chainIDVia(t, NewFallbackTransport([]string{bad.URL, good.URL}))
			if err != nil {
				t.Fatal(err)
			}
			if id != 485590 || badHits != 1 || goodHits != 1 {
				t.Errorf("id=%d badHits=%d goodHits=%d", id, badHits, goodHits)
			}
		})
	}
}

func TestFallbackOnNetworkErrorAndTimeout(t *testing.T) {
	dead := httptest.NewServer(http.NotFoundHandler())
	deadURL := dead.URL
	dead.Close()

	var slowHits, goodHits int32
	slow := rpcServer(t, &slowHits, func(w http.ResponseWriter) bool { time.Sleep(300 * time.Millisecond); return false }, "0x1")
	good := rpcServer(t, &goodHits, nil, "0x768d6")

	ft := NewFallbackTransport([]string{deadURL, slow.URL, good.URL})
	ft.AttemptTimeout = 100 * time.Millisecond
	id, err := chainIDVia(t, ft)
	if err != nil {
		t.Fatal(err)
	}
	if id != 485590 || slowHits != 1 || goodHits != 1 {
		t.Errorf("id=%d slowHits=%d goodHits=%d", id, slowHits, goodHits)
	}
}

func TestNonRateLimitRPCErrorIsReturned(t *testing.T) {
	var badHits, goodHits int32
	bad := rpcServer(t, &badHits, func(w http.ResponseWriter) bool {
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"nonce too low"}}`))
		return true
	}, "")
	good := rpcServer(t, &goodHits, nil, "0x768d6")
	_, err := chainIDVia(t, NewFallbackTransport([]string{bad.URL, good.URL}))
	if err == nil || !strings.Contains(err.Error(), "nonce too low") {
		t.Fatalf("want the node's error, got %v", err)
	}
	if goodHits != 0 {
		t.Errorf("a real JSON-RPC answer must not be re-sent elsewhere, goodHits=%d", goodHits)
	}
}

func TestPreferredURLRetriedAfterCooldown(t *testing.T) {
	var down int32 = 1
	var prefHits, backupHits int32
	pref := rpcServer(t, &prefHits, func(w http.ResponseWriter) bool {
		if atomic.LoadInt32(&down) == 1 {
			w.WriteHeader(502)
			return true
		}
		return false
	}, "0x768d6")
	backup := rpcServer(t, &backupHits, nil, "0x768d6")

	now := time.Unix(1_800_000_000, 0)
	ft := NewFallbackTransport([]string{pref.URL, backup.URL})
	ft.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if _, err := chainIDVia(t, ft); err != nil {
			t.Fatal(err)
		}
	}
	if prefHits != 1 || backupHits != 3 {
		t.Fatalf("during cooldown: prefHits=%d backupHits=%d, want 1 and 3", prefHits, backupHits)
	}

	atomic.StoreInt32(&down, 0)
	now = now.Add(31 * time.Second)
	if _, err := chainIDVia(t, ft); err != nil {
		t.Fatal(err)
	}
	if prefHits != 2 || backupHits != 3 {
		t.Errorf("after cooldown: prefHits=%d backupHits=%d, want 2 and 3", prefHits, backupHits)
	}
}

func TestAllProvidersDownTriesEachOnceSequentially(t *testing.T) {
	var aHits, bHits, inFlight, maxInFlight int32
	fail := func(hits *int32) func(w http.ResponseWriter) bool {
		return func(w http.ResponseWriter) bool {
			n := atomic.AddInt32(&inFlight, 1)
			if n > atomic.LoadInt32(&maxInFlight) {
				atomic.StoreInt32(&maxInFlight, n)
			}
			time.Sleep(20 * time.Millisecond)
			atomic.AddInt32(&inFlight, -1)
			w.WriteHeader(500)
			return true
		}
	}
	a := rpcServer(t, &aHits, fail(&aHits), "")
	b := rpcServer(t, &bHits, fail(&bHits), "")
	ft := NewFallbackTransport([]string{a.URL, b.URL})
	req, _ := http.NewRequest("POST", a.URL, strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"eth_sendRawTransaction","params":["0x00"]}`))
	if _, err := ft.RoundTrip(req); err == nil {
		t.Fatal("want an error when every provider fails")
	}
	if aHits != 1 || bHits != 1 || maxInFlight != 1 {
		t.Errorf("aHits=%d bHits=%d maxInFlight=%d, want 1, 1, 1", aHits, bHits, maxInFlight)
	}
}

func TestRequestBodyReplayed(t *testing.T) {
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := ioutil.ReadAll(r.Body)
		bodies = append(bodies, string(b))
		if len(bodies) == 1 {
			w.WriteHeader(503)
			return
		}
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`))
	}))
	defer srv.Close()
	ft := NewFallbackTransport([]string{srv.URL, srv.URL + "/second"})
	payload := `{"jsonrpc":"2.0","id":1,"method":"eth_chainId","params":[]}`
	req, _ := http.NewRequest("POST", srv.URL, strings.NewReader(payload))
	resp, err := ft.RoundTrip(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if len(bodies) != 2 || bodies[0] != payload || bodies[1] != payload {
		t.Errorf("bodies = %q", bodies)
	}
}
