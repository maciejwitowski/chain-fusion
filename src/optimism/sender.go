package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var keysFolderName = ".keys"
var pkAFilename = "accountA"
var pkBFilename = "accountB"
var faucetFilename = "faucet"

var l1RpcUrl = "http://0.0.0.0:8545"
var l2RpcUrl = "http://0.0.0.0:9545"

// Sends tokens between 2 account on locally running Optimism devnet
func main() {
	err := ensureKeysExists()
	if err != nil {
		log.Fatal("Error while ensuring keys exist:", err)
	}

	//Connect to L2 RPC
	client, err := ethclient.Dial(l2RpcUrl)
	if err != nil {
		log.Fatal("error connecting to L2 RPC:", err)
	}

	privateKey, err := ethcrypto.LoadECDSA(filepath.Join(keysFolderName, pkAFilename))
	if err != nil {
		log.Fatal("error reading priv key: ", err)
	}

	faucetKey, err := ethcrypto.LoadECDSA(filepath.Join(keysFolderName, faucetFilename))
	if err != nil {
		log.Fatal("error reading faucet key: ", err)
	}

	fmt.Println("Fetching balance BEFORE")

	balance, err := client.BalanceAt(context.Background(), ethcrypto.PubkeyToAddress(privateKey.PublicKey), nil)
	if err != nil {
		log.Fatal("error getting balance: ", err)
	} else {
		fmt.Println("account A balance: ", balance)
	}

	balance, err = client.BalanceAt(context.Background(), ethcrypto.PubkeyToAddress(faucetKey.PublicKey), nil)
	if err != nil {
		log.Fatal("error getting balance: ", err)
	} else {
		fmt.Println("faucet balance: ", balance)
	}

	faucet := NewFaucet(client, faucetKey)
	tx, err := faucet.requestFunds(context.Background(), ethcrypto.PubkeyToAddress(privateKey.PublicKey), params.Ether)
	if err != nil {
		log.Fatal("error getting funds from faucet: ", err)
	} else {
		fmt.Println("Request funds tx:", tx.Hash())
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		receipt, err := waitForReceipt(client, tx.Hash())
		if err != nil {
			fmt.Println("error fetching receipt: ", err)
			return
		}
		log.Printf("Transaction mined in block: %d", receipt.BlockNumber.Uint64())
	}()
	wg.Wait()

	fmt.Println("Fetching balance AFTER")

	balance, err = client.BalanceAt(context.Background(), ethcrypto.PubkeyToAddress(privateKey.PublicKey), nil)
	if err != nil {
		log.Fatal("error getting balance: ", err)
	} else {
		fmt.Println("account A balance: ", balance)
	}

	balance, err = client.BalanceAt(context.Background(), ethcrypto.PubkeyToAddress(faucetKey.PublicKey), nil)
	if err != nil {
		log.Fatal("error getting balance: ", err)
	} else {
		fmt.Println("faucet balance: ", balance)
	}
}

// Generate private keys A and B if they don't exist
func ensureKeysExists() error {
	keysFolder, err := KeysFolder()
	if err != nil {
		return err
	}
	if _, err := os.Stat(keysFolder); os.IsNotExist(err) {
		err = os.MkdirAll(keysFolder, 0755)
		if err != nil {
			return err
		}
	}

	pkAPath := filepath.Join(keysFolder, pkAFilename)
	if _, err := os.Stat(pkAPath); os.IsNotExist(err) {
		pk, err := GeneratePrivateKey()
		if err != nil {
			log.Fatal("Error generating PK:", err)
		}
		err = StorePrivateKey(pk, pkAPath)
		if err != nil {
			log.Fatal("Error storing PK:", err)
		}

		fmt.Println("Created PK A")
	}

	pkBPath := filepath.Join(keysFolder, pkBFilename)
	if !fileExists(pkBPath) {
		pk, err := GeneratePrivateKey()
		if err != nil {
			log.Fatal("Error generating PK:", err)
		}
		err = StorePrivateKey(pk, pkBPath)
		if err != nil {
			log.Fatal("Error storing PK:", err)
		}

		fmt.Println("Created PK B")
	}

	fmt.Println("All good with keys!")
	return nil
}

func KeysFolder() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(currentDir, keysFolderName), nil
}

func waitForReceipt(client *ethclient.Client, txHash common.Hash) (*types.Receipt, error) {
	ctx := context.Background()
	for {
		fmt.Println("fetching receipt")
		receipt, err := client.TransactionReceipt(ctx, txHash)
		if err == nil {
			return receipt, nil
		}
		if !errors.Is(err, ethereum.NotFound) {
			return nil, err
		}

		latestBlock, err := client.BlockByNumber(context.Background(), nil)
		if err != nil {
			log.Fatalf("Error getting latest block: %v", err)
		}
		latestBlockNumber := latestBlock.Number()

		block, err := client.BlockByNumber(context.Background(), latestBlockNumber)
		if err != nil {
			log.Fatalf("Error getting block: %v", err)
		}

		transactions := block.Transactions()
		fmt.Println("Number of tx:", len(transactions))

		time.Sleep(time.Second) // Wait for 1 second before checking again
	}
}
