package main
import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
)

/*
Echo server written with websockets.
Server listen for a message on a websocket channel and write it back.
The client is an angular application the uses WebSocket object
in order to exchange messages with backend.

From a browser:  http://localhost:8080/echo.html
*/
func main() {
	fs := http.Dir("web/")
	fileHandler := http.FileServer(fs)
	http.Handle("/", fileHandler)
	http.HandleFunc("/ws", wsHandler)

	log.Println("Running ...")
	err := http.ListenAndServe(":8080", nil)
	log.Println(err.Error())
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// from gorilla website
	conn, err := websocket.Upgrade(w, r, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(w, "Not a websocket handshake", 400)
		return
	} else if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, "Read message from websocket error", 400)
			return
		}
		log.Println(string(msg))
		if err = conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			http.Error(w, "Write message to websocket error", 400)
			return
		}
	}

}