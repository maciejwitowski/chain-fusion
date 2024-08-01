package main

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"os"
)

func fileExists(path string) bool {
	_, err := os.Stat("../keys/privkeyA")
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func GeneratePrivateKey() (string, error) {
	pkA, err := crypto.GenerateKey()
	if err != nil {
		return "", err
	}
	pkBytes := crypto.FromECDSA(pkA)
	return hexutil.Encode(pkBytes)[2:], nil
}

func StorePrivateKey(pk string, path string) error {
	return os.WriteFile(path, []byte(pk), os.FileMode(0644))
}
