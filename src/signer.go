package main

import (
	"context"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

type SignerFn func(context.Context, common.Address, *types.Transaction) (*types.Transaction, error)

func privateKeySignerFn(privateKey *ecdsa.PrivateKey, chainID uint64) SignerFn {
	signer := types.LatestSignerForChainID(big.NewInt(int64(chainID)))

	return func(ctx context.Context, address common.Address, transaction *types.Transaction) (*types.Transaction, error) {
		signature, err := crypto.Sign(signer.Hash(transaction).Bytes(), privateKey)
		if err != nil {
			return nil, err
		}
		res, err := transaction.WithSignature(signer, signature)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

}
