package main
import (
	"log"
	"sync"
)


type hub struct {
	sync.RWMutex
	// registered connections
	connections map[*Connection]bool

	// register new connections
	register chan *Connection

	// unregister unavailable connections
	unregister chan *Connection

	// messages come into hub from here
	// and will be broadcast to all registered connections
	input chan []byte
}

var h = &hub{
	connections: make(map[*Connection]bool),
	register: make(chan *Connection),
	unregister: make(chan *Connection),
	input: make(chan []byte),
}


func (h *hub) run() {
	go func() {
		defer log.Println("Exit from hub run()")

		for {
			select {
			case c := <-h.register:
				log.Println("Registering connection to ", c.ws.RemoteAddr())
				h.connections[c] = true
			case c := <-h.unregister:
				log.Println("Unregistering connection to ", c.ws.RemoteAddr())
				h.removeConnection(c)
			case m, ok := <-h.input:
				if !ok {
					h.unregisterAll()
					return
				}
				for c, _ := range h.connections {
					select {
					case c.send <- m:
					default:
						log.Println("Cannot write to", c.ws.RemoteAddr())
					}
				}
			}
		}
	}()
}


func (h *hub) Registered(c *Connection) bool {
	h.RLock()
	_, ok := h.connections[c]
	h.RUnlock()
	return ok
}

func (h *hub) removeConnection(c *Connection) {
	h.Lock()
	defer h.Unlock()

	_, found := h.connections[c]
	// try to avoid closing an already closed channel
	if found {
		delete(h.connections, c)
		close(c.send)
	}
}

func (h *hub) unregisterAll() {
	log.Println("Unregister all connections.")
	h.Lock()
	defer h.Unlock()

	for c, _ := range h.connections {
		close(c.send)
		delete(h.connections, c)
	}
}