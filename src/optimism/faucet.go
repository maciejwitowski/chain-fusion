package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
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
	faucetAddress := ethcrypto.PubkeyToAddress(f.privateKey.PublicKey)
	nonce, err := f.ethClient.PendingNonceAt(ctx, faucetAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	gasLimit := uint64(21000)
	//tip, err := f.ethClient.SuggestGasTipCap(ctx)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to suggest gas tip cap: %w", err)
	//}

	//gasFeePerGas, err := f.ethClient.SuggestGasPrice(ctx)
	//if err != nil {
	//	return nil, fmt.Errorf("failed to suggest gas price: %w", err)
	//}

	chainId, err := f.ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	} else {
		fmt.Println("chain id:", chainId)
	}

	gasPrice := big.NewInt(20000000000)
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &receiverAddress,
		Value:    new(big.Int).SetUint64(amount),
		Data:     nil,
	})

	signedTx, err := types.SignTx(tx, types.NewEIP2930Signer(chainId), f.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	err = f.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}
	return tx, nil
}
