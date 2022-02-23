package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/btcsuite/btcd/rpcclient"
	"github.com/jmjac/overDrawer/blockchain"
	"github.com/jmjac/overDrawer/btcState"
	"github.com/jmjac/overDrawer/postgres"
	"github.com/jmjac/overDrawer/rest"
	"github.com/jmjac/overDrawer/store"
	"github.com/jmjac/vrscClient"
)

func main() {
	vrscPass := os.Getenv("vrscPass")
	client := os.Getenv("vrscClient")
	vrscAddr := os.Getenv("vrscAddr")

	//setupBTC()
	verus := vrscClient.New(vrscAddr, client, vrscPass)

	var startNew bool
	var serverPort int
	var dbPort int
	var dbBTCPort int
	flag.BoolVar(&startNew, "new", false, "Start the scan from 0. You need to clean the stored blocks before running it")
	flag.IntVar(&serverPort, "p", 8080, "Listening port for the server")
	flag.IntVar(&dbPort, "dbVRSC", 5432, "Port for verus db")
	flag.IntVar(&dbBTCPort, "dbBTC", 5432, "Port for btc db")
	flag.Parse()
	port := ":" + strconv.Itoa(serverPort)

	//btcState, err := setupBTC(dbBTCPort)
	//if err != nil {
	//log.Fatal(err)
	//}
	//btcState.Scan(make(chan bool))
	//return

	//Check if the coin deamon is running
	deamonErrors := 0
	_, err := verus.GetBlockCount()
	for err != nil {
		deamonErrors++
		if deamonErrors > 60 {
			log.Fatal("Couldn't reach deamon in 60 tries. Shutting down")
		}
		log.Println(err)
		log.Println("Sleeping 10s to wait for deamon")
		time.Sleep(time.Second * 10)
		_, err = verus.GetBlockCount()
	}

	dbUser := os.Getenv("dbUser")
	dbPass := os.Getenv("dbPass")
	dbName := os.Getenv("dbName")
	dbHost := os.Getenv("dbHost")

	//Setup the data storage
	psql, err := postgres.Open(dbUser, dbPass, dbName, dbHost, dbPort)
	if err != nil {
		log.Fatalf("MAIN: %v\n", err)
	}
	store := store.New(psql)

	var state blockchain.BlockchainState
	//TODO: Rewrite
	if startNew {
		state = blockchain.New(verus, &store)
	} else {
		state = blockchain.LoadBlockchainState(verus, &store)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Calculating stats")
	log.Println("Identities:", len(state.Identities))

	s := rest.NewServer(port, &verus, &state)
	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func setupVerus() (*blockchain.BlockchainState, error) {
	return nil, nil
}

func setupBTC(dbPort int) (*btcState.BtcState, error) {
	dbUser := os.Getenv("dbBTCUser")
	dbPass := os.Getenv("dbBTCPass")
	dbName := os.Getenv("dbBTCName")
	dbHost := os.Getenv("dbBTCHost")
	btcUser := os.Getenv("BTCUser")
	btcPass := os.Getenv("BTCPass")
	btcHost := os.Getenv("BTCHost")

	//TODO: Change later
	config := &rpcclient.ConnConfig{
		HTTPPostMode: true,
		DisableTLS:   true,
		Host:         btcHost,
		User:         btcUser,
		Pass:         btcPass,
	}
	client, err := rpcclient.New(config, nil)

	deamonErrors := 0
	_, err = client.GetBlockCount()
	for err != nil {
		deamonErrors++
		if deamonErrors > 60 {
			log.Fatal("Couldn't reach deamon in 60 tries. Shutting down")
		}
		log.Println(err)
		log.Println("Sleeping 10s to wait for deamon")
		time.Sleep(time.Second * 10)
		_, err = client.GetBlockCount()
	}

	//Setup the data storage
	psql, err := postgres.Open(dbUser, dbPass, dbName, dbHost, dbPort)
	if err != nil {
		return nil, err
	}
	btcStore := store.New(psql)
	btcState := btcState.LoadBlockchainState(client, &btcStore)
	return &btcState, nil
}
