package main

import "fmt"


func (cli *CLI) addBlock(data string) {

}

func (cli *CLI) createBlockChain(address string) {
	if !isValidAddress(address) {
		fmt.Println("The address is invalid, the invalid address is: ", address)
		return
	}
	err := CreateBlockChain(address)
	if err != nil {
		fmt.Println("CreateBlockChain failed:", err)
		return
	}
	fmt.Println("Finished!")
}

func (cli *CLI) print() {
	bc, err := GetBlockChainInstance()
	if err != nil {
		fmt.Println("print err:", err)
		return
	}
	defer bc.db.Close()
	it := bc.NewIterator()
	for {
		block := it.Next()
		fmt.Printf("\n++++++++++++++++++++++\n")
		fmt.Printf("Version : %d\n", block.Version)
		fmt.Printf("PrevHash : %x\n", block.PrevHash)
		fmt.Printf("MerkleRoot : %x\n", block.MerkleRoot)
		fmt.Printf("TimeStamp : %d\n", block.TimeStamp)
		fmt.Printf("Bits : %d\n", block.Bits)
		fmt.Printf("Nonce : %d\n", block.Nonce)
		fmt.Printf("Hash : %x\n", block.Hash)
		fmt.Printf("Data : %s\n", block.Transactions[0].TXInputs[0].ScriptSig)
		pow := NewProofOfWork(block)
		fmt.Printf("IsValid: %v\n", pow.IsValid())
		if block.PrevHash == nil {
			fmt.Println("Blockchain traversal is over!")
			break
		}
	}
}

func (cli *CLI) getBalance(address string) {
	if !isValidAddress(address) {
		fmt.Println("The address is invalid, the invalid address is: ", address)
		return
	}
	bc, err := GetBlockChainInstance()
	if err != nil {
		fmt.Println("getBalance err:", err)
		return
	}
	defer bc.db.Close()
	pubKeyHash := getPubKeyHashFromAddress(address)
	utxoinfos := bc.FindMyUTXO(pubKeyHash)
	total := 0.0
	for _, utxo := range utxoinfos {
		total += utxo.TXOutput.Value
	}
	fmt.Printf("'%s''s amount is: %f\n", address, total)
}

func (cli *CLI) send(from, to string, amount float64, miner, data string) {
	if !isValidAddress(from) {
        fmt.Println("from is invalid, the invalid address is: ", from)
		return
	}
	if !isValidAddress(to) {
		fmt.Println("to is invalid, the invalid address is: ", to)
		return
	}
	if !isValidAddress(miner) {
		fmt.Println("miner is invalid, the invalid address is: ", miner)
		return
	}
	bc, err := GetBlockChainInstance()
	if err != nil {
		fmt.Println("send err:", err)
		return
	}
	defer bc.db.Close()
	coinbaseTx := NewCoinbaseTx(miner, data)
	txs := []*Transaction{coinbaseTx}
	tx := NewTransaction(from, to, amount, bc)
	if tx != nil {
		fmt.Println("Found a valid transfer transaction!")
		txs = append(txs, tx)
	} else {
		fmt.Println("Note that if an invalid transfer transaction is found, it will not be added to the block!")
	}
	err = bc.AddBlock(txs)
	if err != nil {
		fmt.Println("Failed to add block, transfer failed!")
	}
	fmt.Println("The block is added successfully and the transfer is successful!")
}

func (cli *CLI) createWallet() {
	wm := NewWalletManager()
	if wm == nil {
		fmt.Println("createWallet failed!")
		return
	}
	address := wm.createWallet()
	if len(address) == 0 {
		fmt.Println("Failed to create wallet!")
		return
	}
	fmt.Println("The new wallet address is:", address)
}

func (cli *CLI) listAddress() {
	wm := NewWalletManager()
	if wm == nil {
		fmt.Println(" NewWalletManager failed!")
		return
	}
	addresses := wm.listAddresses()
	for _, address := range addresses {
		fmt.Printf("%s\n", address)
	}
}

func (cli *CLI) printTx() {
	bc, err := GetBlockChainInstance()
	if err != nil {
		fmt.Println("getBalance err:", err)
		return
	}
	defer bc.db.Close()
	it := bc.NewIterator()
	for {
		block := it.Next()
		fmt.Println("\n+++++++++++++++++ block +++++++++++++++")

		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}
}
