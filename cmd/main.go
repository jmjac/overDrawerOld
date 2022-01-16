package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmjac/overDrawer/blockchain"
	"github.com/jmjac/overDrawer/rest"
	"github.com/jmjac/vrscClient"
)

func main() {
	pass := os.Getenv("vrscPass")
	verus := vrscClient.New("client", pass)
	bc, err := verus.GetBlockCount()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(bc)
	state, err := blockchain.LoadBlockchainState("state.json")
	state.SetBlockchain(verus)
	state.GetLockedIdentities()
	//state.SaveToDisk("state.json")

	fmt.Println(state.CalculateStats())
	//state := blockchain.New(verus)

	//TODO: Figure out how to stop this on command
	//	go state.Scan()
	//now := time.Now()
	//fmt.Println(state.TransactionsInLastBlocks(60 * 24 * 30))
	//fmt.Println(time.Since(now))

	//TODO: Put it into an sql databse, instead of memory
	//fmt.Println(len(state.Blocks))
	//state.CreateTransactionHistory()
	//fmt.Println(len(state.Blocks))
	//state.SaveToDisk("texxt.txt")
	//locked := state.GetLockedIdentities()
	//fmt.Println(len(locked))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Identities:", len(state.Identities))
	s := rest.NewServer(":8081", &verus, &state)
	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}

	//identities := make([]string, 0)
	//for i := range state.Identities {
	//identities = append(identities, i)
	//}

	//updateIdentitiesList(verus, identities)

	//ch := make(chan string, 10)
	//go state.Scan(verus, bc, ch)

	//time.Sleep(time.Second * 5)
	//TODO: Change to update as summary every day
	//updateIdentitesAndSendAlert(verus, identities, ch)
}
