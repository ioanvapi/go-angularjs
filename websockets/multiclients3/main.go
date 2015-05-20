package main
import (
	"net/http"
	"log"
	"encoding/json"
	"time"
)

/*
The idea is to send the same content to multiple clients (websocket connections)
that can appear or disappear live

From at least 2 browsers:  http://localhost:8080/client.html
*/

var content = make([]string, 0)


func main() {

	http.HandleFunc("/ws", handler)
	http.Handle("/", http.FileServer(http.Dir("web/")))

	h.run()

	feedSomeData(h)

	log.Println("Running ...")
	err := http.ListenAndServe(":8080", nil)
	log.Println(err.Error())
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := NewConnection(w, r)
	if err != nil {
		http.Error(w, "Not a websocket handshake", 400)
		return
	}
	h.register <- conn
	go conn.StartWriting()
	//todo send some data
}


func feedSomeData(h *hub) {
	go func() {
		tick := time.NewTicker(1 * time.Second)

		for {
			<-tick.C
			if len(h.connections) == 0 {
				continue
			}
			h.input <- getContentAsJSON()
		}
	}()
}


func getContentAsJSON() []byte {
	// append at beginning
	content = append([]string{time.Now().Format("Jan 2 15:04:05")}, content...)
	if len(content) > 25 {
		content = content[:25]
	}
	j, _ := json.Marshal(content)
	return j
}