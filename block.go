package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"time"
)

type Block struct {
	Version uint64
	PrevHash []byte
	MerkleRoot []byte
	TimeStamp uint64
	Bits uint64
	Nonce uint64
	Hash []byte
	Transactions []*Transaction
}

func NewBlock(txs []*Transaction, prevHash []byte) *Block {
	b := Block{
		Version:    0,
		PrevHash:   prevHash,
		MerkleRoot: nil, 
		TimeStamp:  uint64(time.Now().Unix()),
		Bits:  0, 
		Nonce: 0, 
		Hash:  nil,
		Transactions: txs,
	}
	b.HashTransactionMerkleRoot()
	fmt.Printf("merkleRoot:%x\n", b.MerkleRoot)
	pow := NewProofOfWork(&b)
	hash, nonce := pow.Run()
	b.Hash = hash
	b.Nonce = nonce
	return &b
}

func (b *Block) Serialize() []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(b)
	if err != nil {
		fmt.Printf("Encode err:", err)
		return nil
	}
	return buffer.Bytes()
}

func Deserialize(src []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(src))
	err := decoder.Decode(&block)
	if err != nil {
		fmt.Printf("decode err:", err)
		return nil
	}
	return &block
}

func (block *Block) HashTransactionMerkleRoot() {
	var info [][]byte
	for _, tx := range block.Transactions {
		txHashValue := tx.TXID //[]byte
		info = append(info, txHashValue)
	}
	value := bytes.Join(info, []byte{})
	hash := sha256.Sum256(value)
	block.MerkleRoot = hash[:]
}
