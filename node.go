package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	bc "tchain/blockchain"
	nt "tchain/network"
)

func init() {
	if len(os.Args) < 2 {
		panic("failed: len(os.Args) < 2")
	}
	var (
		serveStr     = ""
		addrStr      = ""
		userNewStr   = ""
		userLoadStr  = ""
		chainNewStr  = ""
		chainLoadStr = ""
	)
	var (
		serveExist     = false
		addrExist      = false
		userNewExist   = false
		userLoadExist  = false
		chainLoadExist = false
		chainNewExist  = false
	)

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case strings.HasPrefix(arg, "-serve:"):
			chainLoadStr = strings.Replace(arg, "-serve:", "", 1)
			chainLoadExist = true
		case strings.HasPrefix(arg, "-loadchain:"):
			serveStr = strings.Replace(arg, "-loadchain:", "", 1)
			serveExist = true
		case strings.HasPrefix(arg, "-newchain:"):
			chainNewStr = strings.Replace(arg, "-newchain:", "", 1)
			chainNewExist = true
		case strings.HasPrefix(arg, "-loadaddr:"):
			addrStr = strings.Replace(arg, "-loadaddr:", "", 1)
			addrExist = true
		case strings.HasPrefix(arg, "-newuser:"):
			userNewStr = strings.Replace(arg, "-newuser:", "", 1)
			userNewExist = true
		case strings.HasPrefix(arg, "-loaduser:"):
			userLoadStr = strings.Replace(arg, "-loaduser:", "", 1)
			userLoadExist = true
		}
	}
	if !(userNewExist || userLoadExist) || !addrExist || !serveExist ||
		!(chainNewExist || chainLoadExist) {
		panic("failed: !(userNewExist || userLoadExist) || !addrExist || !serveExist !(chainNewExist || chainLoadExist)")
	}
	Serve = serveStr
	var addresses []string
	err := json.Unmarshal([]byte(readFile(addrStr)), &Addresses)
	if err != nil {
		panic("failed: load addresses")
	}
	var mapaddr = make(map[string]bool)
	for _, addr := range addresses {
		if addr == Serve {
			continue
		}
		if _, ok := mapaddr[addr]; ok {
			continue
		}
		mapaddr[addr] = true
		Addresses = append(Addresses, addr)
	}
	if userNewExist {
		User = userNew(userNewStr)
	}
	if userLoadExist {
		User = userLoad(userLoadStr)
	}
	if User.PrivateKey == nil {
		panic("failed: load user")
	}
	if chainNewExist {
		Filename = chainNewStr
		Chain = chainNew(chainNewStr)
	}
	if chainLoadExist {
		Filename = chainLoadStr
		Chain = chainNew(chainLoadStr)
	}
	if Chain == nil {
		panic("faild 6")
	}
	Block = bc.NewBlock(User.Address(), Chain.LastHash())
}

func main() {
	nt.Listen(Serve, handleServerServe)
	for {
		fmt.Scanln()
	}
}

func chainNew(filename string) *bc.BlockChain {
	bc.NewChain(filename, User.Address())
	return bc.LoadChain(filename)
}

func chainLoad(filename string) *bc.BlockChain {
	chain := bc.LoadChain(filename)
	return chain
}
