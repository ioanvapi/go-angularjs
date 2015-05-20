package main
import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Connection struct {
	// websocket connection to the client
	ws *websocket.Conn

	// channel for messages
	send chan []byte
}


func NewConnection(w http.ResponseWriter, r *http.Request) (*Connection, error) {
	client, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &Connection{
		// websocket connection with the client
		ws: client,
		// here comes messages that should be send on websocket
		send: make(chan ([]byte)),
	}, nil
}

// Should be invoked after NewConnection() is created in order to be active.
// this should be live while client is available
// it can be closed closing the c.send channel
func (c *Connection)StartWriting() {
	defer func() {
		c.ws.Close()
		log.Println("Closing connection to ", c.ws.RemoteAddr())
	}()

	for {
		message, ok := <-c.send
		if !ok {
			c.ws.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			return
		}
	}
}