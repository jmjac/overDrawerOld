package rest

import (
	"log"
	"net/http"
)

func (s Server) logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v requested %v ", r.RemoteAddr, r.URL.Path)
		f(w, r)
	}
}

func (s Server) corsHandler(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://www.overdrawer.com")
		w.Header().Set("Access-Control-Allow-Origin", "http://overdrawer.com")
		w.Header().Set("Access-Control-Allow-Origin", "https://www.overdrawer.com")
		w.Header().Set("Access-Control-Allow-Origin", "https://overdrawer.com")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		f(w, r)
	}
}
