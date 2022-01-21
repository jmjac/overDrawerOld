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
	pass := os.Getenv("vrscPass")
	client := os.Getenv("vrscClient")
	verus := vrscClient.New(client, pass)

	var startNew bool
	var serverPort int
	flag.BoolVar(&startNew, "new", false, "Start the scan from 0. You need to clean the stored blocks before running it")
	flag.IntVar(&serverPort, "p", 8080, "Listening port for the server")
	flag.Parse()
	port := ":" + strconv.Itoa(serverPort)

	//Check if the coin deamon is running
	_, err := verus.GetBlockCount()
	if err != nil {
		log.Fatal(err)
	}

	//TODO: Change to env variable
	user := "overdrawer"
	password := "s9471923uhdasujdh9u12jueh19e2"
	dbname := "overdrawer"

	//Setup the data storage
	psql, err := postgres.Open(user, password, dbname)
	if err != nil {
		log.Fatal(err)
	}
	store := store.New(psql)

	filename := "state2.json"
	var state blockchain.BlockchainState
	if startNew {
		state = blockchain.New(verus, &store, filename)
	} else {
		state, err = blockchain.LoadBlockchainState(filename, verus, &store)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println(store)
	state.SetBlockchain(verus)
	state.CalculateStats()
	c := make(chan bool)
	state.Scan(c)
	state.SaveToDisk()

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Identities:", len(state.Identities))
	s := rest.NewServer(port, &verus, &state)
	err = s.Run()
	if err != nil {
		log.Fatal(err)
	}
}
