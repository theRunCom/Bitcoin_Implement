package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type ProofOfWork struct {
	block *Block
	target *big.Int
}

func NewProofOfWork(block *Block) *ProofOfWork {
	pow := ProofOfWork{
		block: block,
	}
	targetStr := "0001000000000000000000000000000000000000000000000000000000000000"
	tmpBigInt := new(big.Int)
	tmpBigInt.SetString(targetStr, 16)
	pow.target = tmpBigInt
	return &pow
}

func (pow *ProofOfWork) Run() ([]byte, uint64) {
	var nonce uint64
	var hash [32]byte
	fmt.Println("Start mining...")

	for {
		fmt.Printf("%x\r", hash[:])
		data := pow.PrepareData(nonce)
		hash = sha256.Sum256(data)
		tmpInt := new(big.Int)
		tmpInt.SetBytes(hash[:])
		if tmpInt.Cmp(pow.target) == -1 {
			fmt.Printf("Successful mining, hash : %x, nonce : %d\n", hash[:], nonce)
			break
		} else {
			nonce++
		}
	}
	return hash[:], nonce
}

func (pow *ProofOfWork) PrepareData(nonce uint64) []byte {
	b := pow.block
	tmp := [][]byte{
		uintToByte(b.Version), 
		b.PrevHash,
		b.MerkleRoot, 
		uintToByte(b.TimeStamp),
		uintToByte(b.Bits),
		uintToByte(nonce),
	}
	data := bytes.Join(tmp, []byte{})
	return data
}

func (pow *ProofOfWork) IsValid() bool {
	data := pow.PrepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	tmpInt := new(big.Int)
	tmpInt.SetBytes(hash[:])
	return tmpInt.Cmp(pow.target) == -1
}
