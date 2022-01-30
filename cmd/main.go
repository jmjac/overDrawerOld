package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/jmjac/overDrawer/blockchain"
	"github.com/jmjac/overDrawer/postgres"
	"github.com/jmjac/overDrawer/rest"
	"github.com/jmjac/overDrawer/store"
	"github.com/jmjac/vrscClient"
)

func main() {
	vrscPass := os.Getenv("vrscPass")
	client := os.Getenv("vrscClient")
	vrscAddr := os.Getenv("vrscAddr")
	verus := vrscClient.New(vrscAddr, client, vrscPass)

	var startNew bool
	var serverPort int
	var dbPort int
	flag.BoolVar(&startNew, "new", false, "Start the scan from 0. You need to clean the stored blocks before running it")
	flag.IntVar(&serverPort, "p", 8080, "Listening port for the server")
	flag.IntVar(&dbPort, "p", 5432, "Port for db")
	flag.Parse()
	port := ":" + strconv.Itoa(serverPort)

	//Check if the coin deamon is running
	_, err := verus.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}

	dbUser := os.Getenv("dbUser")
	dbPass := os.Getenv("dbPass")
	dbName := os.Getenv("dbName")
	dbHost := os.Getenv("dbHost")

	//Setup the data storage
	psql, err := postgres.Open(dbUser, dbPass, dbName, dbHost, dbPort)
	if err != nil {
		log.Fatal(err)
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

	fmt.Println("Calculating stats")
	log.Println("Identities:", len(state.Identities))

	s := rest.NewServer(port, &verus, &state)
	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
