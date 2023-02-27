package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Transaction struct {
	TXID      []byte     
	TXInputs  []TXInput  
	TXOutputs []TXOutput 
	TimeStamp uint64     
}

type TXInput struct {
	Txid  []byte 
	Index int64  
	ScriptSig []byte 
	PubKey    []byte 
}

type TXOutput struct {
	ScriptPubKeyHash []byte  
	Value            float64 
}

func newTXOutput(address string, amount float64) TXOutput {
	output := TXOutput{Value: amount}
	pubKeyHash := getPubKeyHashFromAddress(address)
	output.ScriptPubKeyHash = pubKeyHash
	return output
}

func (tx *Transaction) setHash() error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		fmt.Println("encode err:", err)
		return err
	}
	hash := sha256.Sum256(buffer.Bytes())
	tx.TXID = hash[:]
	return nil
}

var reward = 12.5

func NewCoinbaseTx(miner string, data string) *Transaction {
	input := TXInput{Txid: nil, Index: -1, ScriptSig: nil, PubKey: []byte(data)}
	output := newTXOutput(miner, reward)
	timeStamp := time.Now().Unix()
	tx := Transaction{
		TXID:      nil,
		TXInputs:  []TXInput{input},
		TXOutputs: []TXOutput{output},
		TimeStamp: uint64(timeStamp),
	}
	tx.setHash()
	return &tx
}

func (tx *Transaction) isCoinbaseTx() bool {
	inputs := tx.TXInputs
	if len(inputs) == 1 && inputs[0].Txid == nil && inputs[0].Index == -1 {
		return true
	}
	return false
}

func NewTransaction(from, to string, amount float64, bc *BlockChain) *Transaction {
	wm := NewWalletManager()
	if wm == nil {
		fmt.Println("Failed to open wallet!")
		return nil
	}

	wallet, ok := wm.Wallets[from]
	if !ok {
		fmt.Println("The private key corresponding to the address was not found!")
		return nil
	}
	fmt.Println("Find the private and public keys of the payer, ready to create the transaction...")
	priKey := wallet.PriKey 
	pubKey := wallet.PubKey
	pubKeyHash := getPubKeyHashFromPubKey(pubKey)
	var spentUTXO = make(map[string][]int64)
	var retValue float64
	spentUTXO, retValue = bc.findNeedUTXO(pubKeyHash, amount)
	if retValue < amount {
		fmt.Println("Insufficient amount, failed to create transaction!")
		return nil
	}
	var inputs []TXInput
	var outputs []TXOutput
	for txid, indexArray := range spentUTXO {
		for _, i := range indexArray {
			input := TXInput{Txid: []byte(txid), Index: i, ScriptSig: nil, PubKey: pubKey}
			inputs = append(inputs, input)
		}
	}
	output1 := newTXOutput(to, amount)
	outputs = append(outputs, output1)
	if retValue > amount {
		output2 := newTXOutput(from, retValue-amount)
		outputs = append(outputs, output2)
	}
	timeStamp := time.Now().Unix()
	tx := Transaction{nil, inputs, outputs, uint64(timeStamp)}
	tx.setHash()
	if !bc.signTransaction(&tx, priKey) {
		fmt.Println("Transaction signing failed")
		return nil
	}
	return &tx
}

func (tx *Transaction) sign(priKey *ecdsa.PrivateKey, prevTxs map[string]*Transaction) bool {
	fmt.Println("Specific to the transaction signature sign...")
	if tx.isCoinbaseTx() {
		fmt.Println("Find mining transactions, no signature required!")
		return true
	}
	txCopy := tx.trimmedCopy()
	for i, input := range txCopy.TXInputs {
		fmt.Printf("input[%d] to sign...\n", i)
		prevTx := prevTxs[string(input.Txid)]
		if prevTx == nil {
			return false
		}
		output := prevTx.TXOutputs[input.Index]
		txCopy.TXInputs[i].PubKey = output.ScriptPubKeyHash
		txCopy.setHash()
		txCopy.TXInputs[i].PubKey = nil
		hashData := txCopy.TXID 
		r, s, err := ecdsa.Sign(rand.Reader, priKey, hashData)
		if err != nil {
			fmt.Println("Signature failed!")
			return false
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.TXInputs[i].ScriptSig = signature
	}
	fmt.Println("Transaction signing successful!")
	return true
}

func (tx *Transaction) trimmedCopy() *Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, input := range tx.TXInputs {
		input := TXInput{
			Txid:      input.Txid,
			Index:     input.Index,
			ScriptSig: nil,
			PubKey:    nil,
		}
		inputs = append(inputs, input)
	}
	outputs = tx.TXOutputs
	txCopy := Transaction{tx.TXID, inputs, outputs, tx.TimeStamp}
	return &txCopy
}

func (tx *Transaction) verify(prevTxs map[string]*Transaction) bool {
	txCopy := tx.trimmedCopy()
	for i, input := range tx.TXInputs {
		prevTx := prevTxs[string(input.Txid)]
		if prevTx == nil {
			return false
		}
		output := prevTx.TXOutputs[input.Index]
		txCopy.TXInputs[i].PubKey = output.ScriptPubKeyHash
		txCopy.setHash()
		txCopy.TXInputs[i].PubKey = nil
		hashData := txCopy.TXID
		signature := input.ScriptSig
		pubKey := input.PubKey
		var r, s, x, y big.Int
		r.SetBytes(signature[:len(signature)/2])
		s.SetBytes(signature[len(signature)/2:])
		x.SetBytes(pubKey[:len(pubKey)/2])
		y.SetBytes(pubKey[len(pubKey)/2:])
		curve := elliptic.P256()
		pubKeyRaw := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		res := ecdsa.Verify(&pubKeyRaw, hashData, &r, &s)
		if !res {
			fmt.Println("An input that failed validation was found!")
			return false
		}
	}
	fmt.Println("Transaction verification successful!")
	return true
}

func (tx *Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.TXID))
	for i, input := range tx.TXInputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Index))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.ScriptSig))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}
	for i, output := range tx.TXOutputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %f", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.ScriptPubKeyHash))
	}
	return strings.Join(lines, "\n")
}
