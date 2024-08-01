package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
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

	privateKeyA, err := ethcrypto.LoadECDSA(filepath.Join(keysFolderName, pkAFilename))
	if err != nil {
		log.Fatal("error reading priv key: ", err)
	}

	faucetKey, err := ethcrypto.LoadECDSA(filepath.Join(keysFolderName, faucetFilename))
	if err != nil {
		log.Fatal("error reading faucet key: ", err)
	}

	fmt.Println("Fetching balance BEFORE")

	addressA := ethcrypto.PubkeyToAddress(privateKeyA.PublicKey)
	balance, err := client.BalanceAt(context.Background(), addressA, nil)
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

	nonce, err := client.NonceAt(context.Background(), faucet.Address(), nil)
	if err != nil {
		return
	}
	fmt.Println("Nonce: ", nonce)

	tx, err := faucet.requestFunds(context.Background(), ethcrypto.PubkeyToAddress(privateKeyA.PublicKey), params.Ether)
	if err != nil {
		log.Fatal("error getting funds from faucet: ", err)
	} else {
		fmt.Println("Request funds tx:", tx.Hash())
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := waitForFunds(context.Background(), client, faucet.Address(), nonce)
		if err != nil {
			fmt.Println("Error while waiting for tx", err)
			return
		} else {
			fmt.Println("Transaction done.")
			return
		}
	}()
	wg.Wait()

	fmt.Println("Fetching balance AFTER")

	balance, err = client.BalanceAt(context.Background(), ethcrypto.PubkeyToAddress(privateKeyA.PublicKey), nil)
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

func waitForFunds(ctx context.Context, client *ethclient.Client, senderAddress common.Address, prevNonce uint64) error {
	for {
		fmt.Println("Waiting for tx...")
		nonceNow, err := client.NonceAt(ctx, senderAddress, nil)
		if err != nil {
			return nil
		}

		fmt.Printf("Nonce before: %d, after: %d\n", prevNonce, nonceNow)
		if nonceNow != prevNonce {
			return nil
		}
		time.Sleep(time.Second)
	}
}
