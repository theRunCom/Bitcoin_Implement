package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

type wallet struct {
	PriKey *ecdsa.PrivateKey
	PubKey []byte
}

func newWalletKeyPair() *wallet {
	curve := elliptic.P256()
	priKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		fmt.Println("ecdsa.GenerateKey err:", err)
		return nil
	}
	pubKeyRaw := priKey.PublicKey
	pubKey := append(pubKeyRaw.X.Bytes(), pubKeyRaw.Y.Bytes()...)
	wallet := wallet{priKey, pubKey}
	return &wallet
}

func (w *wallet) getAddress() string {
	pubKeyHash := getPubKeyHashFromPubKey(w.PubKey)
	payload := append([]byte{byte(0x00)}, pubKeyHash...)
	checksum := checkSum(payload)
	payload = append(payload, checksum...)
	address := base58.Encode(payload)
	return address
}

func getPubKeyHashFromPubKey(pubKey []byte) []byte {
	hash1 := sha256.Sum256(pubKey)
	hasher := ripemd160.New()
	hasher.Write(hash1[:])
	pubKeyHash := hasher.Sum(nil)
	return pubKeyHash
}

func getPubKeyHashFromAddress(address string) []byte {
	decodeInfo := base58.Decode(address)
	if len(decodeInfo) != 25 {
		fmt.Println("getPubKeyHashFromAddress, the address is invalid")
		return nil
	}
	pubKeyHash := decodeInfo[1 : len(decodeInfo)-4]
	return pubKeyHash
}

func checkSum(payload []byte) []byte {
	first := sha256.Sum256(payload)
	second := sha256.Sum256(first[:])
	checksum := second[0:4]
	return checksum
}

func isValidAddress(address string) bool {
	decodeInfo := base58.Decode(address)
	if len(decodeInfo) != 25 {
		fmt.Println("isValidAddress, the length of address is invalid")
		return false
	}
	payload := decodeInfo[:len(decodeInfo)-4]   
	checksum1 := decodeInfo[len(decodeInfo)-4:]
	checksum2 := checkSum(payload)
	return bytes.Equal(checksum1, checksum2)
}
