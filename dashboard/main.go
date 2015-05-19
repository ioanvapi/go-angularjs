package main
import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"encoding/json"
	"github.com/gorilla/mux"
	"sync"
)

var list *EventsList

type Event struct {
	Host string          `json:"host"`
	Service string       `json:"service"`
	State string         `json:"state"`
	Time string          `json:"time"`
	Description string   `json:"description"`
	Tags []string        `json:"tags"`
	Metric string        `json:"metric"`
}

type Handlers struct {
	channels map[*websocket.Conn]chan bool
}

func (h *Handlers)Add(conn *websocket.Conn)(chan bool) {
	ch := make(chan bool)
	h.channels[conn] = ch
	log.Printf("added a channel connection to map: conn %v\n", conn.RemoteAddr())
	return ch
}

func (h *Handlers)Remove(conn *websocket.Conn) {
	delete(h.channels, conn)
	log.Printf("removed a channel connection from map: conn %v\n", conn.RemoteAddr())
}

func main() {
	list = NewEventsList()
	hs := &Handlers{make(map[*websocket.Conn]chan bool)}

	r := mux.NewRouter()
	r.HandleFunc("/ws", hs.wsHandler())
	r.HandleFunc("/api/events", GetEventsHandler).Methods("GET")
	r.HandleFunc("/api/event", hs.PostEventsHandler()).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	http.Handle("/", r)

	log.Println("Running ...")
	err := http.ListenAndServe(":8080", nil)
	log.Println(err.Error())
}

// it responds to a GET request when we are asked for all events
func GetEventsHandler(w http.ResponseWriter, r *http.Request) {
	msg, err := json.Marshal(list.Data())
	if err != nil {
		http.Error(w, "encoding memory list of events error", http.StatusInternalServerError)
		log.Printf("encoding memory list of events error: %v\n", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(msg)
	log.Printf("GetEventsHandler(): '%s'", string(msg))
}

// it responds to a POST request, when we get  a new event
func (h *Handlers)PostEventsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "error decoding event received on 'api'", http.StatusInternalServerError)
			return
		}
		log.Printf("Getting: %+v\n", event)
		list.Add(event)
		for conn, ch := range h.channels {
			log.Printf("Send update signal to %v\n", conn)
			go func() {
				ch <- true
			}()
		}
	}
}

func (h *Handlers)wsHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// from gorilla website
		conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", http.StatusInternalServerError)
			return
		} else if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()
		log.Println("***********  Connected to websocket   *************")

		ch := h.Add(conn)
		defer h.Remove(conn)

		for {
			log.Println("***  Waiting for an internal update ")
			<-ch
			msg, err := json.Marshal(list.Data())
			if err != nil {
				http.Error(w, "decoding data event to json error", http.StatusInternalServerError)
				return
			}

			if err = conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				http.Error(w, "Write message to websocket error", http.StatusInternalServerError)
				return
			}
			log.Printf("wrote to ws\n")
		}
	}
}

type EventsList struct {
	sync.RWMutex
	data []Event
}

func NewEventsList() *EventsList {
	return &EventsList{
		data: make([]Event, 0),
	}
}

func (list *EventsList)Data() []Event {
	return list.data
}

func (list *EventsList)Add(e Event) {

	list.data = append(list.data, e)
}