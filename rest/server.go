package rest

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmjac/overDrawer/blockchain"
	"github.com/jmjac/vrscClient"
)

type Server struct {
	server     *http.Server
	router     *mux.Router
	verus      *vrscClient.Verus
	blockState *blockchain.BlockchainState
	addr       string
	terminate  chan bool
}

func NewServer(addr string, verus *vrscClient.Verus, state *blockchain.BlockchainState) *Server {
	s := &Server{}
	s.addr = addr
	s.router = mux.NewRouter()
	s.verus = verus
	//TODO: Maybe this should start the scan
	s.blockState = state
	s.terminate = make(chan bool)
	//state.GetLockedIdentities()
	go s.blockState.Scan(s.terminate)
	s.server = &http.Server{Addr: addr, Handler: s.router}
	s.routes()
	return s
}

func (s Server) Run() error {
	log.Printf("Starting server at %v\n", s.server.Addr)
	//TODO: Add TSL later
	err := s.server.ListenAndServe()
	return err
}
