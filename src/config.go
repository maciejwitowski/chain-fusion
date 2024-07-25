package main

import (
	"crypto/ecdsa"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	RPCEndpoint string
	PrivateKey  *ecdsa.PrivateKey
}

func InitConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	privateKey, err := ethcrypto.LoadECDSA(".priv_key")
	if err != nil {
		return nil, err
	}

	return &Config{
		RPCEndpoint: os.Getenv("RPC_URL"),
		PrivateKey:  privateKey,
	}, nil
}
