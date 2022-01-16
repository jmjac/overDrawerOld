package rest

func (s Server) routes() {
	s.router.HandleFunc("/identities", s.logging(s.handleIdentities))
	s.router.HandleFunc("/stats", s.logging(s.handleStats))
	s.router.HandleFunc("/identity", s.logging(s.handleIdentity))
	s.router.HandleFunc("/blockcount", s.logging(s.handleBlockCount))
	s.router.HandleFunc("/lockedidentities", s.logging(s.handleLockedIdentities))
}
