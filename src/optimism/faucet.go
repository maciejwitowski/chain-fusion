package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"math/big"
)

type Faucet struct {
	ethClient  *ethclient.Client
	privateKey *ecdsa.PrivateKey // Faucet private key to sign transactions
}

func NewFaucet(client *ethclient.Client, faucetKey *ecdsa.PrivateKey) *Faucet {
	return &Faucet{
		ethClient:  client,
		privateKey: faucetKey,
	}
}

func (f *Faucet) requestFunds(ctx context.Context, receiverAddress common.Address, amount uint64) (*types.Transaction, error) {
	nonce, err := f.ethClient.PendingNonceAt(ctx, f.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasLimit, err := f.ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From:  receiverAddress, // Your sender address
		To:    &receiverAddress,
		Value: new(big.Int).SetUint64(amount),
		Data:  nil,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas: %w", err)
	}

	gasLimit = gasLimit * 120 / 100

	// Get the suggested gas price
	gasFeePerGas, err := f.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to suggest gas price: %w", err)
	}
	tip := new(big.Int).Div(gasFeePerGas, big.NewInt(2))
	gasFeeCap := new(big.Int).Mul(gasFeePerGas, big.NewInt(2))

	chainId, err := f.ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	} else {
		fmt.Println("chain id:", chainId)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &receiverAddress,
		Value:     new(big.Int).SetUint64(amount),
		Data:      nil,
	})

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainId), f.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = f.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}
	return tx, nil
}

func (f *Faucet) Address() common.Address {
	return ethcrypto.PubkeyToAddress(f.privateKey.PublicKey)
}
