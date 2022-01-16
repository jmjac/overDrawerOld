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
