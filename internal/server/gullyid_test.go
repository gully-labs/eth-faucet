package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestSetGullyID(t *testing.T) {
	cases := []struct {
		name, client, issuer   string
		wantClient, wantIssuer string
		wantErr                bool
	}{
		{name: "off by default", client: "", issuer: "https://id.example.xyz"},
		{name: "blank is off", client: "   ", issuer: ""},
		{name: "default issuer", client: "gully-faucet", wantClient: "gully-faucet", wantIssuer: DefaultGullyIDIssuer},
		{name: "trailing slash dropped", client: "gully-faucet", issuer: "https://id.example.xyz/", wantClient: "gully-faucet", wantIssuer: "https://id.example.xyz"},
		{name: "port kept", client: "gully-faucet", issuer: "https://id.example.xyz:8443", wantClient: "gully-faucet", wantIssuer: "https://id.example.xyz:8443"},
		{name: "http localhost", client: "gully-faucet", issuer: "http://localhost:4000", wantClient: "gully-faucet", wantIssuer: "http://localhost:4000"},
		{name: "http elsewhere", client: "gully-faucet", issuer: "http://id.example.xyz", wantErr: true},
		{name: "javascript scheme", client: "gully-faucet", issuer: "javascript:alert(1)", wantErr: true},
		{name: "path", client: "gully-faucet", issuer: "https://id.example.xyz/sdk.js", wantErr: true},
		{name: "query", client: "gully-faucet", issuer: "https://id.example.xyz?x=1", wantErr: true},
		{name: "userinfo", client: "gully-faucet", issuer: "https://a@id.example.xyz", wantErr: true},
		{name: "bad client", client: "Gully Faucet<script>", wantErr: true},
		{name: "client too long", client: "a1234567890123456789012345678901234567890123456789012345678901234", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := NewConfig("testnet", "ETH", 8080, 1440, 0, 1, "", "")
			err := cfg.SetGullyID(tc.client, tc.issuer)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got client=%q issuer=%q", cfg.gullyIDClient, cfg.gullyIDIssuer)
				}
				if cfg.gullyIDClient != "" {
					t.Fatalf("client set despite error: %q", cfg.gullyIDClient)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.gullyIDClient != tc.wantClient || cfg.gullyIDIssuer != tc.wantIssuer {
				t.Fatalf("got client=%q issuer=%q, want %q %q", cfg.gullyIDClient, cfg.gullyIDIssuer, tc.wantClient, tc.wantIssuer)
			}
		})
	}
}

func infoJSON(t *testing.T, cfg *Config) map[string]interface{} {
	t.Helper()
	mockBuilder := new(MockTxBuilder)
	mockBuilder.On("Sender").Return(common.HexToAddress("0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045"))
	srv := NewServer(mockBuilder, cfg)
	req := httptest.NewRequest("GET", "/api/info", nil)
	rr := httptest.NewRecorder()
	srv.handleInfo().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d", rr.Code)
	}
	var out map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestHandleInfoGullyIDOff(t *testing.T) {
	out := infoJSON(t, NewConfig("testnet", "ETH", 8080, 1440, 0, 1, "", ""))
	if _, ok := out["gullyid_client"]; ok {
		t.Fatalf("gullyid_client present while off: %v", out)
	}
	if _, ok := out["gullyid_issuer"]; ok {
		t.Fatalf("gullyid_issuer present while off: %v", out)
	}
}

func TestHandleInfoGullyIDOn(t *testing.T) {
	cfg := NewConfig("testnet", "ETH", 8080, 1440, 0, 1, "", "")
	if err := cfg.SetGullyID("gully-faucet", ""); err != nil {
		t.Fatal(err)
	}
	out := infoJSON(t, cfg)
	if out["gullyid_client"] != "gully-faucet" {
		t.Fatalf("gullyid_client = %v", out["gullyid_client"])
	}
	if out["gullyid_issuer"] != DefaultGullyIDIssuer {
		t.Fatalf("gullyid_issuer = %v", out["gullyid_issuer"])
	}
}
