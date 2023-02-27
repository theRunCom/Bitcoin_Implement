package main

import (
	"fmt"
	"os"
	"strconv"
)

type CLI struct {

}

const Usage = `
The Usage of 
	./blockchain create <ADDRESS> 
	./blockchain addBlock <ADD INFO> 
	./blockchain print
	./blockchain getBalance <ADDRESS>
	./blockchain send <FROM> <TO> <AMOUNT> <MINER> <DATA>
	./blockchain createWallet
	./blockchain listAddress
	./blockchain printTx
`

func (cli *CLI) Run() {
	cmds := os.Args
	if len(cmds) < 2 {
		fmt.Println("Invalid input parameter, please check!")
		fmt.Println(Usage)
		return
	}
	switch cmds[1] {
	case "create":
		fmt.Println("Create block command called!")
		if len(cmds) != 3 {
			fmt.Println("Invalid input parameter, please check!")
			return
		}
		address := cmds[2]
		cli.createBlockChain(address)
	case "addBlock":
		if len(cmds) != 3 {
			fmt.Println("Invalid input parameter, please check!")
			return
		}
		data := cmds[2]
		cli.addBlock(data)
	case "print":
		fmt.Println("Print block command called!")
		cli.print()
	case "getBalance":
		fmt.Println("Get balance command called!")
		if len(cmds) != 3 {
			fmt.Println("Invalid input parameter, please check!")
			return
		}
		address := cmds[2] //需要检验个数
		cli.getBalance(address)
	case "send":
		fmt.Println("Send command called")
		if len(cmds) != 7 {
			fmt.Println("Invalid input parameter, please check!")
			return
		}
		from := cmds[2]
		to := cmds[3]
		amount, _ := strconv.ParseFloat(cmds[4], 64)
		miner := cmds[5]
		data := cmds[6]
		cli.send(from, to, amount, miner, data)
	case "createWallet":
		fmt.Println("Createwallet command called")
		cli.createWallet()
	case "listAddress":
		fmt.Println("Listaddress command called")
		cli.listAddress()
	case "printTx":
		cli.printTx()
	default:
		fmt.Println("Invalid input parameter, please check!")
		fmt.Println(Usage)
	}
}
