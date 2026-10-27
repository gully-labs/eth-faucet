package chain

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
)

const (
	// fallbackGasLimit is used when eth_estimateGas fails: enough for a plain transfer on Arbitrum Nitro, where the
	// L1 data fee is charged as extra gas on top of the 21000 of an EVM transfer.
	fallbackGasLimit = uint64(60000)
	// gasLimitMarginPct is applied to the node's estimate: the L1 data component moves with the parent chain's gas
	// price between the estimate and the block that includes the transfer.
	gasLimitMarginPct = uint64(130)
)

// gasLimitFor turns an eth_estimateGas result into the gas limit of the transfer: the estimate plus a 30% margin, or
// the fallback when the estimate failed. A fixed 21000 fails on Arbitrum Nitro with "intrinsic gas too low".
func gasLimitFor(estimate uint64, err error) uint64 {
	if err != nil || estimate == 0 {
		return fallbackGasLimit
	}
	return estimate * gasLimitMarginPct / 100
}

type TxBuilder interface {
	Sender() common.Address
	Transfer(ctx context.Context, to string, value *big.Int) (common.Hash, error)
}

type TxBuild struct {
	client          bind.ContractTransactor
	privateKey      *ecdsa.PrivateKey
	signer          types.Signer
	fromAddress     common.Address
	nonce           uint64
	supportsEIP1559 bool
}

// NewTxBuilder connects to provider, a single RPC URL or a comma-separated ordered list of http(s) URLs (empty =
// DefaultProviders). With http(s) URLs every call goes through a FallbackTransport: first healthy URL wins, the next
// one is tried after a network error, timeout, 5xx/429 or rate-limit error.
func NewTxBuilder(provider string, privateKey *ecdsa.PrivateKey, chainID *big.Int) (TxBuilder, error) {
	client, err := dialProviders(ParseProviders(provider))
	if err != nil {
		return nil, err
	}

	if chainID == nil {
		chainID, err = client.ChainID(context.Background())
		if err != nil {
			return nil, err
		}
	}

	supportsEIP1559, err := checkEIP1559Support(client)
	if err != nil {
		return nil, err
	}

	txBuilder := &TxBuild{
		client:          client,
		privateKey:      privateKey,
		signer:          types.NewLondonSigner(chainID),
		fromAddress:     crypto.PubkeyToAddress(privateKey.PublicKey),
		supportsEIP1559: supportsEIP1559,
	}
	txBuilder.refreshNonce(context.Background())

	return txBuilder, nil
}

func (b *TxBuild) Sender() common.Address {
	return b.fromAddress
}

func (b *TxBuild) Transfer(ctx context.Context, to string, value *big.Int) (common.Hash, error) {
	if value == nil || value.Sign() <= 0 {
		return common.Hash{}, fmt.Errorf("invalid transfer value: must be positive")
	}

	toAddress := common.HexToAddress(to)
	estimate, estimateErr := b.client.EstimateGas(ctx, ethereum.CallMsg{From: b.fromAddress, To: &toAddress, Value: value})
	if estimateErr != nil {
		log.WithFields(log.Fields{"error": estimateErr, "to": toAddress}).Warn("gas estimate failed, using the fallback gas limit")
	}
	gasLimit := gasLimitFor(estimate, estimateErr)

	nonce := b.getAndIncrementNonce()

	var err error
	var unsignedTx *types.Transaction

	if b.supportsEIP1559 {
		unsignedTx, err = b.buildEIP1559Tx(ctx, &toAddress, value, gasLimit, nonce)
	} else {
		unsignedTx, err = b.buildLegacyTx(ctx, &toAddress, value, gasLimit, nonce)
	}

	if err != nil {
		return common.Hash{}, err
	}

	signedTx, err := types.SignTx(unsignedTx, b.signer, b.privateKey)
	if err != nil {
		return common.Hash{}, err
	}

	if err = b.client.SendTransaction(ctx, signedTx); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "nonce") {
			b.refreshNonce(ctx)
		}
		return common.Hash{}, err
	}

	return signedTx.Hash(), nil
}

func (b *TxBuild) buildEIP1559Tx(ctx context.Context, to *common.Address, value *big.Int, gasLimit uint64, nonce uint64) (*types.Transaction, error) {
	header, err := b.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	gasTipCap, err := b.client.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}

	// gasFeeCap = baseFee * 2 + gasTipCap
	gasFeeCap := new(big.Int).Mul(header.BaseFee, big.NewInt(2))
	gasFeeCap = new(big.Int).Add(gasFeeCap, gasTipCap)

	return types.NewTx(&types.DynamicFeeTx{
		ChainID:   b.signer.ChainID(),
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        to,
		Value:     value,
	}), nil
}

func (b *TxBuild) buildLegacyTx(ctx context.Context, to *common.Address, value *big.Int, gasLimit uint64, nonce uint64) (*types.Transaction, error) {
	gasPrice, err := b.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	return types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       to,
		Value:    value,
	}), nil
}

func (b *TxBuild) getAndIncrementNonce() uint64 {
	return atomic.AddUint64(&b.nonce, 1) - 1
}

func (b *TxBuild) refreshNonce(ctx context.Context) {
	nonce, err := b.client.PendingNonceAt(ctx, b.Sender())
	if err != nil {
		log.WithFields(log.Fields{
			"address": b.Sender(),
			"error":   err,
		}).Error("failed to refresh account nonce")
		return
	}

	atomic.StoreUint64(&b.nonce, nonce)
	log.WithField("nonce", nonce).Info("Nonce refreshed successfully")
}

func dialProviders(urls []string) (*ethclient.Client, error) {
	for _, u := range urls {
		lower := strings.ToLower(u)
		if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
			// ws:// or IPC: no HTTP fallback, use the first URL as before.
			if len(urls) > 1 {
				log.WithField("rpc", u).Warn("non-http provider in a list: only the first URL is used")
			}
			return ethclient.Dial(urls[0])
		}
	}
	log.WithField("providers", strings.Join(urls, ", ")).Info("rpc providers (in fallback order)")
	rpcClient, err := rpc.DialHTTPWithClient(urls[0], &http.Client{Transport: NewFallbackTransport(urls)})
	if err != nil {
		return nil, err
	}
	return ethclient.NewClient(rpcClient), nil
}

func checkEIP1559Support(client *ethclient.Client) (bool, error) {
	header, err := client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return false, err
	}

	return header.BaseFee != nil && header.BaseFee.Cmp(big.NewInt(0)) > 0, nil
}
