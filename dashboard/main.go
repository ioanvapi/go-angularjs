package main
import (
	"net/http"
	"log"
	"encoding/json"
	"github.com/gorilla/mux"
	"time"
	"bytes"
	"os"
	"gitlab.optymyze.net/tools/websocket-hub"
)

// in memory image of the dashboard events
var dashboardEvents *DashboardEvents


func main() {
	db, err := NewBoltStore("./dashboard.db")
	if err != nil {
		log.Fatalln("Cannot create a Bolt store.")
		os.Exit(1)
	}

	if dashboardEvents, err = NewDashboardEvents(db); err != nil {
		log.Println("Error when running NewDashboardEvents(): ", err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/ws", WsEventsHandler)
	r.HandleFunc("/api/events", GetActiveEventsHandler).Methods("GET")
	r.HandleFunc("/api/ackevents", GetAckEventsHandler).Methods("GET")
	r.HandleFunc("/api/event", PostEventHandler).Methods("POST")
	r.HandleFunc("/api/ackevent", PostAckEventHandler).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
	http.Handle("/", r)

	log.Println("Running ...")
	err = http.ListenAndServe(":8080", nil)
	log.Println(err.Error())
}


// WebSocket connection handler. It creates a connection to the WebSocket client
// and starts a goroutine that listen to the connection channel and pushes
// fresh data to the websocket client.
func WsEventsHandler(w http.ResponseWriter, r *http.Request) {
	err := wshub.StartConnection(w, r)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
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
	wshub.Hub.Input <- dashboardEvents.DataJSON()
}

// It responds to a POST request, when we get a new event.
// We update the events list and eventually push new data to dashboard
func PostEventHandler(w http.ResponseWriter, r *http.Request) {
	var event Event
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		// do not respond with http.Error() but with a message about error
		w.Write(error2json(err))
		log.Println("error reading incomming event: ", err)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &event)
	if err != nil {
		// do not respond with http.Error() but with a message about error
		w.Write(error2json(err))
		log.Println("error decoding received event: ", buf.String())
		return
	}

	updated := dashboardEvents.Update(event)
	if updated {
		wshub.Hub.Input <- dashboardEvents.DataJSON()
	}
}

func error2json(err error) []byte {
	var e = struct {error string}{err.Error()}
	json, _ := json.Marshal(e)
	return json
}
