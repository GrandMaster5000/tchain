package main

import (
	"fmt"
	bc "tchain/blockchain"
)

const (
	DBNAME = "blockchain.db"
)

func main() {
	miner := bc.NewUser()
	bc.NewChain(DBNAME, miner.Address())
	chain := bc.LoadChain(DBNAME)
	fmt.Println(chain)
	for i := 0; i < 3; i++ {
		fmt.Println(miner.Address())
		block := bc.NewBlock(miner.Address(), chain.LastHash())
		block.AddTransaction(chain, bc.NewTransaction(miner, chain.LastHash(), "aaa", 5))
		block.AddTransaction(chain, bc.NewTransaction(miner, chain.LastHash(), "bbb", 2))
		block.Accept(chain, miner, make(chan bool))
		chain.AddBlock(block)
	}
	var sblock string
	rows, err := chain.DB.Query("SELECT Block FROM BlockChain")
	if err != nil {
		fmt.Println(err)
		return
	}
	for rows.Next() {
		rows.Scan(&sblock)
		fmt.Println(sblock)
	}
}
