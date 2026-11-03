package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLegalPagesRedirectToGullyChain(t *testing.T) {
	router := http.NewServeMux()
	registerLegalRedirects(router)
	cases := map[string]string{
		"/terms":     "https://gullychain.xyz/terms",
		"/terms/":    "https://gullychain.xyz/terms",
		"/privacy":   "https://gullychain.xyz/privacy",
		"/privacy/":  "https://gullychain.xyz/privacy",
		"/terms?x=1": "https://gullychain.xyz/terms",
	}
	for path, want := range cases {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusMovedPermanently {
			t.Errorf("%s: status %d, want 301", path, rec.Code)
		}
		if got := rec.Header().Get("Location"); got != want {
			t.Errorf("%s: Location %q, want %q", path, got, want)
		}
	}
}

func TestSetupRouterRedirectsLegalPagesAndStillServesTheApp(t *testing.T) {
	cfg := &Config{httpPort: 8080, interval: 1440, network: "testnet", symbol: "ETH", payout: 0.0005}
	if err := cfg.SetProtection(Protection{}); err != nil {
		t.Fatal(err)
	}
	router := NewServer(new(MockTxBuilder), cfg).setupRouter()
	for path, want := range map[string]string{"/terms": "https://gullychain.xyz/terms", "/privacy": "https://gullychain.xyz/privacy"} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusMovedPermanently || rec.Header().Get("Location") != want {
			t.Errorf("%s: %d %q, want 301 %q", path, rec.Code, rec.Header().Get("Location"), want)
		}
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("/: %d, want 200 (the static app)", rec.Code)
	}
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/terms-of-use", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("/terms-of-use: %d, want 404 (only the exact legal paths redirect)", rec.Code)
	}
}
