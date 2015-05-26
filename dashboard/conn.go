package main
import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
	"log"
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

	// interval we must receive connection heartbeats otherwise close the connection
	pingPeriod time.Duration
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
		// heartbeat ping period
		pingPeriod: 40 * time.Second,
	}, nil
}

func (c *Connection) Close() {
	if h.Registered(c) {
		h.unregister <- c
		c.ws.Close()
		log.Println("Closing websocket connection to ", c.ws.RemoteAddr())
	}
}


// Should be invoked after NewConnection() is created in order to be active.
// this should be live while client is available
// it can be closed closing the c.send channel
func (c *Connection)StartWriting() {
	defer c.Close()

	for {
		message, ok := <-c.send
		if !ok {
			log.Println("Reading from closed channel to", c.ws.RemoteAddr())
			c.ws.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Error writing to websocket", c.ws.RemoteAddr())
			return
		}
	}
}


// We expect to receive ping message from dashboard every c.pingPeriod.
// If not, we consider dashboard is closed and we have to unregister and close the connection.
func (c *Connection) ReadPings() {
	defer c.Close()

	// Sets the deadline for future Read calls. After that time, reads will return error
	// and we consider client/dashboard is not alive and we must unregister the connection.
	c.ws.SetReadDeadline(time.Now().Add(c.pingPeriod))

	// consume ping messages
	for {
		//		messageType, data, err := c.ws.ReadMessage()
		_, _, err := c.ws.ReadMessage()
		if err != nil {
			log.Printf("Close connection to '%v' because of a read error in ReadPings(): %v\n", c.ws.RemoteAddr(), err)
			break
		}

		// sets a new deadline for next read
		c.ws.SetReadDeadline(time.Now().Add(c.pingPeriod))
		//		log.Printf("Got message type: '%d' and data: '%s'\n", messageType, string(data))
	}
}