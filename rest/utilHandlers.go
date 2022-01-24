package rest

import (
	"log"
	"net/http"
)

func (s Server) logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("SERVER: %v requested %v ", r.RemoteAddr, r.URL.Path)
		f(w, r)
	}
}

func (s Server) corsHandler(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Content-Type", "application/json")
		f(w, r)
	}
}
