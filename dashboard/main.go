package main
import (
	"net/http"
	"log"
	"encoding/json"
	"github.com/gorilla/mux"
	"time"
)

// in memory image of the dashboard events
var dashboardEvents *DashboardEvents


func main() {
	dashboardEvents = NewDashboardEvents()

	r := mux.NewRouter()
	r.HandleFunc("/ws", WsEventsHandler)
	r.HandleFunc("/api/events", GetActiveEventsHandler).Methods("GET")
	r.HandleFunc("/api/ackevents", GetAckEventsHandler).Methods("GET")
	r.HandleFunc("/api/event", PostEventHandler).Methods("POST")
	r.HandleFunc("/api/ackevent", PostAckEventHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	http.Handle("/", r)

	h.run()

	log.Println("Running ...")
	err := http.ListenAndServe(":8080", nil)
	log.Println(err.Error())
}

// WebSocket connection handler. It creates a connection to the WebSocket client
// and starts a goroutine that listen to the connection channel and pushes
// fresh data to the websocket client.
func WsEventsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := NewConnection(w, r)
	if err != nil {
		http.Error(w, "Not a websocket handshake", 400)
		return
	}
	h.register <- conn
	go conn.StartWriting()
	conn.ReadPings()
}

// It responds to a GET REST request when we are asked for all events
// Using it when dashboard page is first loaded.
func GetActiveEventsHandler(w http.ResponseWriter, r *http.Request) {
	msg := dashboardEvents.ActiveEventsDataJSON()
	w.Header().Set("Content-Type", "application/json")
	w.Write(msg)
//	log.Println("GetEventsHandler()", string(msg))
}

// It responds to a GET REST request when we are asked for all ack events
// Using it when dashboard page is first loaded.
func GetAckEventsHandler(w http.ResponseWriter, r *http.Request) {
	msg := dashboardEvents.AckEventsDataJSON()
	w.Header().Set("Content-Type", "application/json")
	w.Write(msg)
	log.Println("GetAckEventsHandler()")
}

// Invoked as a POST REST request when an event is acknowledged.
// After decode the request Body we have the user who ack and a message.
//
func PostAckEventHandler(w http.ResponseWriter, r *http.Request) {
	var ackEvent Event
	if err := json.NewDecoder(r.Body).Decode(&ackEvent); err != nil {
		w.Write(error2json(err))
		log.Println("error decoding received ACK event: ", err)
		return
	}
	ackEvent.AckTime = time.Now().Unix()
	log.Println("Received ACK:", ackEvent)
	dashboardEvents.MoveActive2AckEvents(ackEvent)
	h.input <- dashboardEvents.DataJSON()
}

// It responds to a POST request, when we get a new event.
// We update the events list and eventually push new data to dashboard
func PostEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		// do not respond with http.Error() but with a message about error
		w.Write(error2json(err))
		log.Println("error decoding received event: ", err)
		return
	}
	updated := dashboardEvents.Update(event)
	//	log.Printf("\n*****************    Size : %+v\n\n", len(list.data))
	if updated {
		h.input <- dashboardEvents.DataJSON()
//		log.Printf("Update based on event:\n %+v\n", event)
//		fmt.Println("Receive at: ", time.Now().Format("15:04:05"))
	}
}

func error2json(err error) []byte {
	var e = struct {error string}{err.Error()}
	json, _ := json.Marshal(e)
	return json
}
