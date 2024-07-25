package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"time"
)

func Run(ctx context.Context, cfg *Config) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Println("Starting the app...")

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	client, err := ethclient.DialContext(ctx, cfg.RPCEndpoint)
	if err != nil {
		return err
	}

	errorChan := make(chan error, 1)
	resultChan := make(chan *big.Int, 1)

	go func() {
		balance, err := client.BalanceAt(ctx, crypto.PubkeyToAddress(cfg.PrivateKey.PublicKey), nil)
		if err == nil {
			resultChan <- balance
		} else {
			errorChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case <-ticker.C:
		log.Println("Timed out")
		return nil
	case balance := <-resultChan:
		log.Printf("Balance: %s", balance.String())
		return nil
	case err = <-errorChan:
		if err != nil {
			return fmt.Errorf("error in goroutine: %w", err)
		}
		return nil
	}
}
