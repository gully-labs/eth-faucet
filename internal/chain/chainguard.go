package chain

import (
	"context"
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

// GullyTestnetChainID is Gully Chain Testnet, the only chain this faucet pays out on by default.
const GullyTestnetChainID = 485590

// DefaultAllowedChainIDs is the start-up allowlist when -faucet.chainids / FAUCET_ALLOWED_CHAIN_IDS is empty.
var DefaultAllowedChainIDs = []int64{GullyTestnetChainID}

// NeverFaucetChainIDs are real-money chains. A faucet hot wallet must never pay out on them, so they are refused even
// when they appear on the allowlist: Gully Chain mainnet 48559, its parent Arbitrum One, and the other mainnets Gully
// products touch.
var NeverFaucetChainIDs = map[int64]string{
	48559: "Gully Chain mainnet",
	1:     "Ethereum",
	42161: "Arbitrum One",
	42170: "Arbitrum Nova",
	8453:  "Base",
	10:    "OP Mainnet",
	56:    "BNB Chain",
	137:   "Polygon",
}

// ParseChainIDs reads a comma-separated list of decimal chain ids. Empty gives DefaultAllowedChainIDs. Mainnets
// (NeverFaucetChainIDs) are an error, so a typo in the env cannot open the faucet on a real-money chain.
func ParseChainIDs(value string) ([]int64, error) {
	var ids []int64
	for _, part := range strings.Split(value, ",") {
		p := strings.TrimSpace(part)
		if p == "" {
			continue
		}
		id, err := strconv.ParseInt(p, 10, 64)
		if err != nil || id <= 0 {
			return nil, fmt.Errorf("allowed chain id %q is not a positive decimal number", p)
		}
		if name, bad := NeverFaucetChainIDs[id]; bad {
			return nil, fmt.Errorf("chain %d (%s) is a mainnet: the faucet never pays out there", id, name)
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return append([]int64(nil), DefaultAllowedChainIDs...), nil
	}
	return ids, nil
}

// CheckChainID refuses a chain id that is not on the allowlist, and every mainnet whatever the allowlist says.
func CheckChainID(id *big.Int, allowed []int64) error {
	if id == nil {
		return fmt.Errorf("the RPC did not report a chain id")
	}
	if id.IsInt64() {
		if name, bad := NeverFaucetChainIDs[id.Int64()]; bad {
			return fmt.Errorf("the RPC is chain %s (%s): the faucet never pays out on a mainnet", id, name)
		}
		for _, a := range allowed {
			if id.Int64() == a {
				return nil
			}
		}
	}
	return fmt.Errorf("the RPC is chain %s, not on the allowed list %v (set FAUCET_ALLOWED_CHAIN_IDS to change it)", id, allowed)
}

// ChainIDReader is the part of an RPC client the guard needs (ethclient.Client implements it).
type ChainIDReader interface {
	ChainID(ctx context.Context) (*big.Int, error)
}

// VerifyChain asks the RPC for its chain id and checks it against the allowlist. A configured chain id (the upstream
// -faucet.name sepolia/holesky shortcut) must also match what the RPC reports, so a signer for one chain is never
// pointed at another. It returns the chain id to sign for.
func VerifyChain(ctx context.Context, client ChainIDReader, configured *big.Int, allowed []int64) (*big.Int, error) {
	got, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("cannot read the chain id from the RPC: %w", err)
	}
	if configured != nil && configured.Cmp(got) != 0 {
		return nil, fmt.Errorf("the RPC is chain %s but the faucet is configured for chain %s", got, configured)
	}
	if err := CheckChainID(got, allowed); err != nil {
		return nil, err
	}
	return got, nil
}
