package main
import "log"


type hub struct {
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
				h.connections[c] = true
				log.Println("Registering connection to ", c.ws.RemoteAddr())
			case c := <-h.unregister:
				delete(h.connections, c)
				close(c.send)
				log.Println("Unregistering connection to ", c.ws.RemoteAddr())
			case m, ok := <-h.input:
				if !ok {
					h.unregisterAll()
					return
				}
				for c, _ := range h.connections {
					select {
					case c.send <- m:
					default:
						close(c.send)
						delete(h.connections, c)
						log.Println("*** Remove connection to ", c.ws.RemoteAddr())
					}
				}
			}
		}
	}()
}

func (h *hub) unregisterAll() {
	for c, _ := range h.connections {
		close(c.send)
		delete(h.connections, c)
	}
}