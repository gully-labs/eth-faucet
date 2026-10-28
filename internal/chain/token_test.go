package chain

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func TestERC20TransferData(t *testing.T) {
	got := common.Bytes2Hex(ERC20TransferData(common.HexToAddress("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"), big.NewInt(1000000)))
	want := "a9059cbb000000000000000000000000ab5801a7d398351b8be11c439e05c5b3259aec9b00000000000000000000000000000000000000000000000000000000000f4240"
	if got != want {
		t.Fatalf("calldata\n got %s\nwant %s", got, want)
	}
}

func TestWithDataKeepsFields(t *testing.T) {
	to := common.HexToAddress("0xDBE84bE6496Fb3d07F601548555ec7Acc4CeBa8F")
	in := types.NewTx(&types.DynamicFeeTx{ChainID: big.NewInt(485590), Nonce: 7, GasTipCap: big.NewInt(1), GasFeeCap: big.NewInt(2), Gas: 90000, To: &to, Value: big.NewInt(0)})
	out := withData(in, []byte{1, 2, 3})
	if out.Nonce() != 7 || out.Gas() != 90000 || *out.To() != to || out.ChainId().Int64() != 485590 || len(out.Data()) != 3 || out.Type() != types.DynamicFeeTxType {
		t.Fatalf("dynamic: %+v", out)
	}
	leg := withData(types.NewTx(&types.LegacyTx{Nonce: 3, GasPrice: big.NewInt(5), Gas: 80000, To: &to, Value: big.NewInt(0)}), []byte{9})
	if leg.Nonce() != 3 || leg.GasPrice().Int64() != 5 || len(leg.Data()) != 1 {
		t.Fatalf("legacy: %+v", leg)
	}
}
