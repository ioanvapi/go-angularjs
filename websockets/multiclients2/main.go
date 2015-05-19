package main
import (
    "net/http"
    "log"
    "github.com/gorilla/websocket"
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
    clients map[*websocket.Conn]bool
    input chan []byte
    register chan *websocket.Conn
    unregister chan *websocket.Conn
}


var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
}

func main() {
    pool := NewWebSocketPool()

    http.HandleFunc("/ws", pool.handler())
    http.Handle("/", http.FileServer(http.Dir("web/")))

    pool.run()

    feedSomeData(pool)

    log.Println("Running ...")
    err := http.ListenAndServe(":8080", nil)
    log.Println(err.Error())
}

func feedSomeData(pool *WebSocketPool) {
    go func() {
        tick := time.NewTicker(5 * time.Second)

        for {
            <-tick.C
            if len(pool.clients) == 0 {
                continue
            }

            if content, err := getContentAsJSON(); err == nil {
                pool.input <- content
            } else {
                log.Printf("Error decoding content: %v\n", err)
            }
        }
    }()
}


func NewWebSocketPool() *WebSocketPool {
    return &WebSocketPool{
        clients: make(map[*websocket.Conn]bool),
        input: make(chan []byte),
        register: make(chan *websocket.Conn),
        unregister: make(chan *websocket.Conn),
    }
}

func (pool *WebSocketPool) handler() func(http.ResponseWriter, *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        client, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
            http.Error(w, "Not a websocket handshake", 400)
            return
        }
        pool.register <- client
        //todo send some data
    }
}

func (pool *WebSocketPool) run() {
    go func() {
        for {
            select {
            case msg, ok := <-pool.input:
                if !ok {
                    log.Println("Data channel is closed. Exit from run().")
                    // close clients
                    for k, _ := range pool.clients {
                        k.Close()
                    }
                    return
                }

                for client, _ := range pool.clients {
                    if err := client.WriteMessage(websocket.TextMessage, msg); err != nil {
                        log.Printf("Error sending message to %v\n", client.RemoteAddr())
                        pool.remove(client)
                        continue
                    }
                    log.Printf("Sent message to %v\n", client.RemoteAddr())
                }
            case client := <-pool.register:
                pool.clients[client] = true
                log.Printf("Register client %v\n", client.RemoteAddr())
            case client := <-pool.unregister:
                pool.remove(client)
            }
        }
    }()
}

func (pool *WebSocketPool) remove(client *websocket.Conn) {
    delete(pool.clients, client)
    client.Close()
    log.Printf("Unregister client %v\n", client.RemoteAddr())
}

// it stops the input channel that in turn close() all the clients connections
// ... then run() stops
func (pool *WebSocketPool)stop() {
    close(pool.input)
}


func getContentAsJSON() ([]byte, error) {
    content = append(content, time.Now().Format("Jan 2 15:04:05"))
    return json.Marshal(content)
}