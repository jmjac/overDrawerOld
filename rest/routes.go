package rest

func (s Server) routes() {
	s.router.HandleFunc("/api/identities", s.logging(s.corsHandler(s.handleIdentities)))
	s.router.HandleFunc("/api/stats", s.logging(s.corsHandler(s.handleStats)))
	s.router.HandleFunc("/api/hourly", s.logging(s.corsHandler(s.handleHourly)))
	s.router.HandleFunc("/api/daily", s.logging(s.corsHandler(s.handleDaily)))
	s.router.HandleFunc("/api/identity", s.logging(s.corsHandler(s.handleIdentity)))
	s.router.HandleFunc("/api/blockcount", s.logging(s.corsHandler(s.handleBlockCount)))
	s.router.HandleFunc("/api/blockHash", s.logging(s.corsHandler(s.handleBlockHash)))
	s.router.HandleFunc("/api/lockedidentities", s.logging(s.corsHandler(s.handleLockedIdentities)))
	//s.router.HandleFunc("/stop", s.logging(s.handleStop))
}
