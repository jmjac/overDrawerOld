package rest

func (s Server) routes() {
	s.router.HandleFunc("/identities", s.logging(s.handleIdentities))
	s.router.HandleFunc("/stats", s.logging(s.handleStats))
	s.router.HandleFunc("/hourly", s.logging(s.handleHourly))
	s.router.HandleFunc("/daily", s.logging(s.handleDaily))
	s.router.HandleFunc("/identity", s.logging(s.handleIdentity))
	s.router.HandleFunc("/blockcount", s.logging(s.handleBlockCount))
	s.router.HandleFunc("/blockHash", s.logging(s.handleBlockHash))
	s.router.HandleFunc("/lockedidentities", s.logging(s.handleLockedIdentities))
	//s.router.HandleFunc("/stop", s.logging(s.handleStop))
}
