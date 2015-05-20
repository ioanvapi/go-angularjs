package main
import (
	"net/http"
	"log"
	"encoding/json"
	"github.com/gorilla/mux"
)

var list *EventsList


func main() {
	list = NewEventsList()

	r := mux.NewRouter()
	r.HandleFunc("/ws", WsHandler)
	r.HandleFunc("/api/events", GetEventsHandler).Methods("GET")
	r.HandleFunc("/api/event", PostEventsHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	http.Handle("/", r)

	h.run()

	log.Println("Running ...")
	err := http.ListenAndServe(":8080", nil)
	log.Println(err.Error())
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := NewConnection(w, r)
	if err != nil {
		http.Error(w, "Not a websocket handshake", 400)
		return
	}
	h.register <- conn
	go conn.StartWriting()
}

// it responds to a GET request when we are asked for all events
func GetEventsHandler(w http.ResponseWriter, r *http.Request) {
	msg := list.DataJSON()
	w.Header().Set("Content-Type", "application/json")
	w.Write(msg)
	log.Printf("GetEventsHandler(): '%s'", string(msg))
}

// it responds to a POST request, when we get  a new event
func PostEventsHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "error decoding event received on 'api'", http.StatusInternalServerError)
		return
	}
	log.Printf("Getting: %+v\n", event)
	list.Add(event)
	h.input <- list.DataJSON()
}

