package main

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
)

type BlockChain struct {
	db   *bolt.DB 
	tail []byte  
}

const genesisInfo = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
const blockchainDBFile = "blockchain.db"
const bucketBlock = "bucketBlock"           
const lastBlockHashKey = "lastBlockHashKey"

func CreateBlockChain(address string) error {
	if isFileExist(blockchainDBFile) {
		fmt.Println("The file is existed!")
		return nil
	}
	db, err := bolt.Open(blockchainDBFile, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketBlock))

		if bucket == nil {
			bucket, err := tx.CreateBucket([]byte(bucketBlock))
			if err != nil {
				return err
			}

			coinbase := NewCoinbaseTx(address, genesisInfo)
			txs := []*Transaction{coinbase}
			genesisBlock := NewBlock(txs, nil)
			bucket.Put(genesisBlock.Hash, genesisBlock.Serialize()) 
			bucket.Put([]byte(lastBlockHashKey), genesisBlock.Hash)
		}
		return nil
	})
	return err 
}

func GetBlockChainInstance() (*BlockChain, error) {
	if isFileExist(blockchainDBFile) == false {
		return nil, errors.New("The flie is not existed, please create it!")
	}

	var lastHash []byte 

	db, err := bolt.Open(blockchainDBFile, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketBlock))
		if bucket == nil {
			return errors.New("bucket shouldn't be nil")
		} else {
			lastHash = bucket.Get([]byte(lastBlockHashKey))
		}
		return nil
	})

	bc := BlockChain{db, lastHash}
	return &bc, nil
}

func (bc *BlockChain) AddBlock(txs1 []*Transaction) error {
	txs := []*Transaction{}

	fmt.Println("Verify the transaction before adding the block...")
	for _, tx := range txs1 {
		if bc.verifyTransaction(tx) {
			fmt.Printf("The current transaction verification is successful: %x\n", tx.TXID)
			txs = append(txs, tx)
		} else {
			fmt.Printf("The current transaction verification failed: %x\n", tx.TXID)
		}
	}

	lashBlockHash := bc.tail 

	newBlock := NewBlock(txs, lashBlockHash)

	err := bc.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketBlock))
		if bucket == nil {
			return errors.New("Bucket shouldn't be nil when adding the block...")
		}

		bucket.Put(newBlock.Hash, newBlock.Serialize())
		bucket.Put([]byte(lastBlockHashKey), newBlock.Hash)

		bc.tail = newBlock.Hash
		return nil
	})

	return err
}

type Iterator struct {
	db          *bolt.DB
	currentHash []byte 
}

func (bc *BlockChain) NewIterator() *Iterator {
	it := Iterator{
		db:          bc.db,
		currentHash: bc.tail,
	}
	return &it
}

func (it *Iterator) Next() (block *Block) {
	err := it.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketBlock))
		if bucket == nil {
			return errors.New("Bucket shouldn't be nil when Iterator Next")
		}
		blockTmpInfo := bucket.Get(it.currentHash)
		block = Deserialize(blockTmpInfo)
		it.currentHash = block.PrevHash
		return nil
	})

	if err != nil {
		fmt.Printf("iterator next err:", err)
		return nil
	}
	return
}

type UTXOInfo struct {
	Txid []byte
	Index int64
	TXOutput
}

func (bc *BlockChain) FindMyUTXO(pubKeyHash []byte) []UTXOInfo {
	var utxoInfos []UTXOInfo
	spentUtxos := make(map[string][]int)
	it := bc.NewIterator()
	for {
		block := it.Next()
		for _, tx := range block.Transactions {
		LABEL:
			for outputIndex, output := range tx.TXOutputs {
				fmt.Println("outputIndex:", outputIndex)
				if bytes.Equal(output.ScriptPubKeyHash, pubKeyHash) {
					currentTxid := string(tx.TXID)
					indexArray := spentUtxos[currentTxid]
					if len(indexArray) != 0 {
						for _, spendIndex := range indexArray {
							if outputIndex == spendIndex {
								continue LABEL
							}
						}
					}
					utxoinfo := UTXOInfo{tx.TXID, int64(outputIndex), output}
					utxoInfos = append(utxoInfos, utxoinfo)
				}
			}

			if tx.isCoinbaseTx() {
				fmt.Println("Discover mining transactions")
				continue
			}

			for _, input := range tx.TXInputs {
				if bytes.Equal(getPubKeyHashFromPubKey(input.PubKey), pubKeyHash) {
					spentKey := string(input.Txid)
					spentUtxos[spentKey] = append(spentUtxos[spentKey], int(input.Index))
				}
			}

		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return utxoInfos
}

func (bc *BlockChain) findNeedUTXO(pubKeyHash []byte, amount float64) (map[string][]int64, float64) {
	var retMap = make(map[string][]int64)
	var retValue float64
	utxoInfos := bc.FindMyUTXO(pubKeyHash)
	for _, utxoinfo := range utxoInfos {
		retValue += utxoinfo.Value
		key := string(utxoinfo.Txid)
		retMap[key] = append(retMap[key], utxoinfo.Index)
		if retValue >= amount {
			break
		}
	}
	return retMap, retValue
}

func (bc *BlockChain) signTransaction(tx *Transaction, priKey *ecdsa.PrivateKey) bool {
	fmt.Println("signTransaction start!!!")
	prevTxs := make(map[string]*Transaction)
	for _, input := range tx.TXInputs {
		prevTx := bc.findTransaction(input.Txid)
		if prevTx == nil {
			fmt.Println("No valid referenced transactions found")
			return false
		}
		fmt.Println("The referenced transaction was found")
		prevTxs[string(input.Txid)] = prevTx
	}
	return tx.sign(priKey, prevTxs)
}

func (bc *BlockChain) verifyTransaction(tx *Transaction) bool {
	fmt.Println("verifyTransaction start!!!")
	if tx.isCoinbaseTx() {
		fmt.Println("Discover mining transactions")
		return true
	}
	prevTxs := make(map[string]*Transaction)
	for _, input := range tx.TXInputs {
		prevTx := bc.findTransaction(input.Txid)
		if prevTx == nil {
			fmt.Println("No valid referenced transactions found")
			return false
		}
		fmt.Println("The referenced transaction was found")
		prevTxs[string(input.Txid)] = prevTx
	}
	return tx.verify(prevTxs)
}

func (bc *BlockChain) findTransaction(txid []byte) *Transaction {
	it := bc.NewIterator()
	for {
		block := it.Next()
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.TXID, txid) {
				return tx
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return nil
}
