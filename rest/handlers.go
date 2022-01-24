package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/jmjac/vrscClient"
)

func (s Server) handleIdentities(w http.ResponseWriter, r *http.Request) {
	identities := make([]string, 0)
	//TODO: Change to be for separate links only
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Content-Type", "application/json")
	out, _ := json.Marshal(identities)
	w.Write(out)
}

func (s Server) handleIdentity(w http.ResponseWriter, r *http.Request) {
	//TODO: Should be get, change in JS too
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST,OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "http://overdrawer.com")
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	type temp struct {
		Identity string
	}

	decoder := json.NewDecoder(r.Body)
	var t temp
	err := decoder.Decode(&t)
	//log.Println(t.Identity)
	identity := t.Identity
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//TODO: This limit may be too low for some identities, check in documentation
	if len(identity) == 0 || len(identity) > 25 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if identity[len(identity)-1] != '@' {
		identity += "@"
	}

	id, err := s.verus.GetIdentity(identity)
	if err != nil {
		if err.Error() == "Identity not found" {
			w.WriteHeader(http.StatusNoContent)
			w.Write([]byte("{}"))
		} else {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	out, _ := json.Marshal(id)
	w.Write(out)
}

func (s Server) handleLockedIdentities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "http://overdrawer.com")
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s.blockState.LockedIdentities)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func (s Server) handleBlockCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "http://overdrawer.com")
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s.blockState.Height)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func (s Server) handleBlockHash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s.blockState.BlockHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}
func (s Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s.blockState.Stats)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func (s Server) handleHourly(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "http://overdrawer.com")
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s.blockState.StatsPerHour)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}
func (s Server) handleDaily(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	//w.Header().Set("Access-Control-Allow-Origin", "http://overdrawer.com")
	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(s.blockState.StatsPerDay)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func (s Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.terminate <- true
}

func updateIdentitesAndSendAlert(verus vrscClient.Verus, identities []string, ch chan string) ([]string, error) {
	readingValues := true
	for readingValues {
		select {
		case idName, ok := <-ch:
			if ok {
				identities = append(identities, idName)
			}
		default:
			readingValues = false
		}
	}
	return identities, nil
}
