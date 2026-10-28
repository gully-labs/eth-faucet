package chain

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	log "github.com/sirupsen/logrus"
)

// TokenSender is the ERC-20 side of the faucet (the test-stablecoin drip next to the ETH payout). TxBuild implements
// it; the server type-asserts its TxBuilder, so builders without it simply have no token drip.
type TokenSender interface {
	TokenBalance(ctx context.Context, token common.Address, owner common.Address) (*big.Int, error)
	TokenDecimals(ctx context.Context, token common.Address) (uint8, error)
	TransferToken(ctx context.Context, token common.Address, to string, amount *big.Int) (common.Hash, error)
}

// fallbackTokenGasLimit is used when eth_estimateGas fails for an ERC-20 transfer: a cold transfer is about 52k EVM
// gas, plus the L1 data component Nitro adds as gas.
const fallbackTokenGasLimit = uint64(150000)

var (
	selBalanceOf = common.FromHex("0x70a08231") // balanceOf(address)
	selDecimals  = common.FromHex("0x313ce567") // decimals()
	selTransfer  = common.FromHex("0xa9059cbb") // transfer(address,uint256)
)

func pad32(b []byte) []byte { return common.LeftPadBytes(b, 32) }

// ERC20TransferData is the calldata of transfer(to, amount).
func ERC20TransferData(to common.Address, amount *big.Int) []byte {
	out := append([]byte{}, selTransfer...)
	out = append(out, pad32(to.Bytes())...)
	return append(out, pad32(amount.Bytes())...)
}

func (b *TxBuild) caller() (bind.ContractCaller, error) {
	c, ok := b.client.(bind.ContractCaller)
	if !ok {
		return nil, errors.New("rpc client cannot make eth_call")
	}
	return c, nil
}

func (b *TxBuild) call(ctx context.Context, token common.Address, data []byte) ([]byte, error) {
	c, err := b.caller()
	if err != nil {
		return nil, err
	}
	out, err := c.CallContract(ctx, ethereum.CallMsg{To: &token, Data: data}, nil)
	if err != nil {
		return nil, err
	}
	if len(out) < 32 {
		return nil, fmt.Errorf("token %s: short eth_call result (%d bytes; not an ERC-20?)", token.Hex(), len(out))
	}
	return out, nil
}

func (b *TxBuild) TokenBalance(ctx context.Context, token common.Address, owner common.Address) (*big.Int, error) {
	out, err := b.call(ctx, token, append(append([]byte{}, selBalanceOf...), pad32(owner.Bytes())...))
	if err != nil {
		return nil, err
	}
	return new(big.Int).SetBytes(out[:32]), nil
}

func (b *TxBuild) TokenDecimals(ctx context.Context, token common.Address) (uint8, error) {
	out, err := b.call(ctx, token, selDecimals)
	if err != nil {
		return 0, err
	}
	d := new(big.Int).SetBytes(out[:32])
	if !d.IsUint64() || d.Uint64() > 36 {
		return 0, fmt.Errorf("token %s: implausible decimals %s", token.Hex(), d)
	}
	return uint8(d.Uint64()), nil
}

// TransferToken sends amount of an ERC-20 from the faucet wallet. It shares the nonce counter with Transfer.
func (b *TxBuild) TransferToken(ctx context.Context, token common.Address, to string, amount *big.Int) (common.Hash, error) {
	if amount == nil || amount.Sign() <= 0 {
		return common.Hash{}, fmt.Errorf("invalid token amount: must be positive")
	}
	data := ERC20TransferData(common.HexToAddress(to), amount)
	estimate, err := b.client.EstimateGas(ctx, ethereum.CallMsg{From: b.fromAddress, To: &token, Data: data})
	if err != nil {
		// a revert here usually means the faucet wallet has no tokens left: do not burn a nonce on it
		if strings.Contains(strings.ToLower(err.Error()), "revert") {
			return common.Hash{}, fmt.Errorf("token transfer would revert: %w", err)
		}
		log.WithFields(log.Fields{"error": err, "token": token}).Warn("token gas estimate failed, using the fallback gas limit")
	}
	gasLimit := fallbackTokenGasLimit
	if err == nil && estimate > 0 {
		gasLimit = estimate * gasLimitMarginPct / 100
	}

	nonce := b.getAndIncrementNonce()
	var unsigned *types.Transaction
	if b.supportsEIP1559 {
		unsigned, err = b.buildEIP1559Tx(ctx, &token, big.NewInt(0), gasLimit, nonce)
	} else {
		unsigned, err = b.buildLegacyTx(ctx, &token, big.NewInt(0), gasLimit, nonce)
	}
	if err != nil {
		b.refreshNonce(ctx)
		return common.Hash{}, err
	}
	unsigned = withData(unsigned, data)
	signed, err := types.SignTx(unsigned, b.signer, b.privateKey)
	if err != nil {
		b.refreshNonce(ctx)
		return common.Hash{}, err
	}
	if err = b.client.SendTransaction(ctx, signed); err != nil {
		b.refreshNonce(ctx)
		return common.Hash{}, err
	}
	return signed.Hash(), nil
}

// withData returns a copy of tx with the given calldata (the builders above only make value transfers).
func withData(tx *types.Transaction, data []byte) *types.Transaction {
	switch tx.Type() {
	case types.DynamicFeeTxType:
		return types.NewTx(&types.DynamicFeeTx{ChainID: tx.ChainId(), Nonce: tx.Nonce(), GasTipCap: tx.GasTipCap(), GasFeeCap: tx.GasFeeCap(), Gas: tx.Gas(), To: tx.To(), Value: tx.Value(), Data: data})
	default:
		return types.NewTx(&types.LegacyTx{Nonce: tx.Nonce(), GasPrice: tx.GasPrice(), Gas: tx.Gas(), To: tx.To(), Value: tx.Value(), Data: data})
	}
}
