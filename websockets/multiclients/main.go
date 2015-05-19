package main
import (
	"net/http"
	"log"
	"github.com/gorilla/websocket"
	"sync"
	"encoding/json"
	"time"
)

/*
The idea is to send the same content to multiple clients (websocket connections)
that can appear or disappear live

From at least 2 browsers:  http://localhost:8080/client.html
*/

var content = make([]string, 0)

type WebSocketPool struct {
	sync.RWMutex
	clients map[*websocket.Conn]int
	input chan []byte
}


var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func main() {
	pool := NewWebSocketPool()

	http.HandleFunc("/ws", pool.handler())
	http.Handle("/", http.FileServer(http.Dir("web/")))

	input := pool.run()

	go feedSomeData(input)

	log.Println("Running ...")
	err := http.ListenAndServe(":8080", nil)
	log.Println(err.Error())
}

func feedSomeData(input chan<- []byte) {
	tick := time.NewTicker(5 * time.Second)

	for {
		<-tick.C
		if content, err := getContentAsJSON(); err == nil {
			input <- content
		} else {
			log.Printf("Error decoding content: %v\n", err)
		}
	}
}


func NewWebSocketPool() *WebSocketPool {
	return &WebSocketPool{
		clients: make(map[*websocket.Conn]int),
	}
}

func (pool *WebSocketPool) handler() func (http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// from gorilla website
		ws, err := websocket.Upgrade(w, r, nil, 1024, 1024)
		if _, ok := err.(websocket.HandshakeError); ok {
			http.Error(w, "Not a websocket handshake", 400)
			return
		} else if err != nil {
			log.Println(err)
			return
		}

		pool.addClient(ws)
		//todo send some data
	}
}

func (pool *WebSocketPool) run() chan<- []byte {
	pool.input = make(chan []byte)

	go func() {
		for {
			msg, ok := <- pool.input
			if !ok {
				log.Println("Data channel is closed. Exit from run().")
				pool.closeClients()
				return
			}

			clients := pool.copyClients()
			for _, client := range clients {
				if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
					pool.deleteClient(client)
					client.Close()
					log.Printf("Error sending message to %v\n", client.RemoteAddr())
					continue
				}
				log.Printf("Sent message to %v\n", client.RemoteAddr())
			}
		}
	}()

	return pool.input
}

// it stops the input channel that in turn close() all the clients connections
// ... then run() stops
func (pool *WebSocketPool)stop() {
	close(pool.input)
}

func (pool *WebSocketPool)closeClients() {
	pool.RLock()
	defer pool.RUnlock()

	for k, _ := range pool.clients {
		k.Close()
	}
}

func (pool *WebSocketPool)copyClients() []*websocket.Conn {
	clients := make([]*websocket.Conn, 0)

	pool.RLock()
	for k, _ := range pool.clients {
		clients = append(clients, k)
	}
	pool.RUnlock()
	return clients
}

func (pool *WebSocketPool)addClient(client *websocket.Conn) {
	pool.Lock()
	pool.clients[client] = 0
	pool.Unlock()
	log.Printf("Added client: %v\n", client.RemoteAddr())
}

func (pool *WebSocketPool)deleteClient(client *websocket.Conn) {
	pool.Lock()
	delete(pool.clients, client)
	pool.Unlock()
	log.Printf("Removed client: %v\n", client.RemoteAddr())
}

func getContentAsJSON() ([]byte, error) {
	content = append(content, time.Now().Format("Jan 2 15:04:05"))
	return json.Marshal(content)
}