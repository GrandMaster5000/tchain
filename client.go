package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	bc "tchain/blockchain"
	nt "tchain/network"
)

func init() {
	if len(os.Args) < 2 {
		panic("failed: len(os.Args) < 2")
	}
	var (
		addrStr     = ""
		userNewStr  = ""
		userLoadStr = ""
	)
	var (
		addrExist     = false
		userNewExist  = false
		userLoadExist = false
	)

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
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
	if !(userNewExist || userLoadExist) || !addrExist {
		panic("falid 2")
	}
	err := json.Unmarshal([]byte(readFile(addrStr)), &Addresses)
	if err != nil {
		panic("failed: load addresses")
	}
	if len(Addresses) == 0 {
		panic("failed: len(Addresses) == 0")
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
}

func main() {
	handleClient()
}

func handleClient() {
	var (
		message string
		splited []string
	)
	for {
		message = inputString("> ")
		splited = strings.Split(message, " ")
		switch splited[0] {
		case "/exit":
			os.Exit(0)
		case "/user":
			if len(splited) < 2 {
				fmt.Println("len(user) < 2")
				continue
			}
			switch splited[1] {
			case "address":
				useAddress()
			case "purse":
				userPurse()
			case "balance":
				userBalance()
			}
		case "/chain":
			if len(splited) < 2 {
				fmt.Println("len(chain) < 2")
				continue
			}
			switch splited[1] {
			case "print":
				chainPrint()
			case "tx":
				chainTX(splited[1:])
			case "balance":
				chainBalance(splited[1:])
			}
		default:
			fmt.Println("undefined command\n")
		}
	}
}

func inputString(begin string) string {
	fmt.Print(begin)
	msg, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	return strings.Replace(msg, "\n", "", 1)
}

func useAddress() {
	fmt.Println("Address:", User.Address(), "\n")
}

func userPurse() {
	fmt.Println("Purse:", User.Purse(), "\n")
}

func userBalance() {
	printBalance(User.Address())
}

func chainPrint() {
	for i := 0; ; i++ {
		res := nt.Send(Addresses[0], &nt.Package{
			Option: GET_BLOCK,
			Data:   fmt.Sprintf("%d", i),
		})
		if res == nil || res.Data == "" {
			break
		}
		fmt.Printf("[%d] => %s\n", i+1, res.Data)
	}
	fmt.Println()
}

func chainTX(splited []string) {
	if len(splited) != 3 {
		fmt.Println("len(splited) != 3 \n")
		return
	}
	num, err := strconv.Atoi(splited[2])
	if err != nil {
		fmt.Println("strconv error \n")
	}
	for _, addr := range Addresses {
		res := nt.Send(addr, &nt.Package{
			Option: GET_LHASH,
		})
		if res == nil {
			continue
		}
		tx := bc.NewTransaction(User, bc.Base64Decode(res.Data), splited[1], uint64(num))
		if tx == nil {
			fmt.Println("tx is null \n")
			break
		}
		res = nt.Send(addr, &nt.Package{
			Option: ADD_TRNSX,
			Data:   bc.SerializeTX(tx),
		})
		if res == nil {
			continue
		}
		if res.Data == "ok" {
			fmt.Printf("ok: (%s)\n", addr)
		} else {
			fmt.Printf("fail: (%s)\n", addr)
		}
	}
	fmt.Println()
}

func chainBalance(splited []string) {
	if len(splited) != 2 {
		fmt.Println("len(splited) != 2\n")
		return
	}
	printBalance(splited[1])
}

func printBalance(address string) {
	for _, addr := range Addresses {
		res := nt.Send(addr, &nt.Package{
			Option: GET_BALANCE,
			Data:   address,
		})
		if res == nil {
			continue
		}
		fmt.Printf("Balnce (%s): %s coins\n", addr, res.Data)
	}
	fmt.Println()
}
