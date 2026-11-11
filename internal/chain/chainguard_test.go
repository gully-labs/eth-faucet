package chain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
)

type fakeChain struct {
	id  *big.Int
	err error
}

func (f fakeChain) ChainID(context.Context) (*big.Int, error) { return f.id, f.err }

func TestParseChainIDsDefaultsToGullyTestnet(t *testing.T) {
	for _, v := range []string{"", " ", ",", " , "} {
		ids, err := ParseChainIDs(v)
		if err != nil || len(ids) != 1 || ids[0] != 485590 {
			t.Fatalf("ParseChainIDs(%q) = %v, %v; want [485590]", v, ids, err)
		}
	}
	ids, err := ParseChainIDs("485590, 421614")
	if err != nil || len(ids) != 2 || ids[1] != 421614 {
		t.Fatalf("list: %v %v", ids, err)
	}
}

func TestParseChainIDsRefusesMainnetsAndJunk(t *testing.T) {
	for _, v := range []string{"48559", "485590,48559", "42161", "1", "8453", "abc", "-5", "0", "0x768d6"} {
		if _, err := ParseChainIDs(v); err == nil {
			t.Fatalf("ParseChainIDs(%q) accepted", v)
		}
	}
}

func TestCheckChainID(t *testing.T) {
	allowed := []int64{485590}
	if err := CheckChainID(big.NewInt(485590), allowed); err != nil {
		t.Fatalf("testnet refused: %v", err)
	}
	cases := map[int64]string{48559: "mainnet", 42161: "mainnet", 421614: "not on the allowed list", 11155111: "not on the allowed list"}
	for id, want := range cases {
		err := CheckChainID(big.NewInt(id), allowed)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Fatalf("chain %d: err %v, want %q", id, err, want)
		}
	}
	// a mainnet stays refused even if someone puts it on the list by hand (bypassing ParseChainIDs)
	if err := CheckChainID(big.NewInt(48559), []int64{48559}); err == nil {
		t.Fatal("48559 accepted from the allowlist")
	}
	huge, _ := new(big.Int).SetString("123456789012345678901234567890", 10)
	if err := CheckChainID(huge, allowed); err == nil {
		t.Fatal("a chain id beyond int64 was accepted")
	}
	if err := CheckChainID(nil, allowed); err == nil {
		t.Fatal("nil chain id accepted")
	}
}

func TestVerifyChain(t *testing.T) {
	ctx := context.Background()
	got, err := VerifyChain(ctx, fakeChain{id: big.NewInt(485590)}, nil, DefaultAllowedChainIDs)
	if err != nil || got.Int64() != 485590 {
		t.Fatalf("testnet: %v %v", got, err)
	}
	if _, err := VerifyChain(ctx, fakeChain{id: big.NewInt(48559)}, nil, DefaultAllowedChainIDs); err == nil {
		t.Fatal("mainnet RPC accepted")
	}
	// configured id (the upstream sepolia shortcut) must match the RPC
	if _, err := VerifyChain(ctx, fakeChain{id: big.NewInt(485590)}, big.NewInt(11155111), []int64{485590, 11155111}); err == nil || !strings.Contains(err.Error(), "configured for chain 11155111") {
		t.Fatalf("mismatch not refused: %v", err)
	}
	if _, err := VerifyChain(ctx, fakeChain{err: errors.New("down")}, nil, DefaultAllowedChainIDs); err == nil {
		t.Fatal("RPC error ignored")
	}
}

// chainIDServer answers eth_chainId with id; every other method gets a JSON-RPC error (the guard must stop before them).
func chainIDServer(t *testing.T, id int64) (*httptest.Server, *[]string) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		calls = append(calls, req.Method)
		w.Header().Set("Content-Type", "application/json")
		if req.Method == "eth_chainId" {
			fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"result":"0x%x"}`, req.ID, id)
			return
		}
		fmt.Fprintf(w, `{"jsonrpc":"2.0","id":%s,"error":{"code":-32601,"message":"not in this test"}}`, req.ID)
	}))
	t.Cleanup(srv.Close)
	return srv, &calls
}

func TestNewTxBuilderRefusesAMainnetRPC(t *testing.T) {
	key, _ := crypto.HexToECDSA("976f9f7772781ff6d1c93941129d417c49a209c674056a3cf5e27e225ee55fa8")
	for _, id := range []int64{48559, 42161, 421614} {
		srv, calls := chainIDServer(t, id)
		b, err := NewTxBuilder(srv.URL, key, nil, DefaultAllowedChainIDs)
		if err == nil || b != nil {
			t.Fatalf("chain %d: builder created", id)
		}
		for _, m := range *calls {
			if m != "eth_chainId" {
				t.Fatalf("chain %d: called %s after a refused chain id", id, m)
			}
		}
	}
	// a configured mainnet id cannot bypass the guard, even when it is also on the list
	srv, _ := chainIDServer(t, 48559)
	if _, err := NewTxBuilder(srv.URL, key, big.NewInt(48559), []int64{48559}); err == nil {
		t.Fatal("configured mainnet id accepted")
	}
}
