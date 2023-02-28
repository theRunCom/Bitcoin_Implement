package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"sort"
)

const walletFile = "wallet.dat"
type WalletManager struct {
	Wallets map[string]*wallet
}

func NewWalletManager() *WalletManager {
	var wm WalletManager
	wm.Wallets = make(map[string]*wallet)
	if !wm.loadFile() {
		return nil
	}
	return &wm
}

func (wm *WalletManager) createWallet() string {
	w := newWalletKeyPair()
	if w == nil {
		fmt.Println("newWalletKeyPair Failed")
		return ""
	}
	address := w.getAddress()
	wm.Wallets[address] = w
	if !wm.saveFile() {
		return ""
	}
	return address
}

func (wm *WalletManager) saveFile() bool {
	var buffer bytes.Buffer
	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(wm)
	if err != nil {
		fmt.Println("encoder.Encode err:", err)
		return false
	}
	err = ioutil.WriteFile(walletFile, buffer.Bytes(), 0600)
	if err != nil {
		fmt.Println("ioutil.WriteFile err:", err)
		return false
	}
	return true
}

func (wm *WalletManager) loadFile() bool {
	if !isFileExist(walletFile) {
		fmt.Println("The file isn't existed, not reload!")
		return true
	}
	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		fmt.Println("ioutil.ReadFile err:", err)
		return false
	}
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))
	err = decoder.Decode(wm)
	if err != nil {
		fmt.Println("decoder.Decode err:", err)
		return false
	}
	return true
}

func (wm *WalletManager) listAddresses() []string {
	var addresses []string
	for address := range wm.Wallets {
		addresses = append(addresses, address)
	}
	sort.Strings(addresses)
	return addresses
}
