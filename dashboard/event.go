package main
import (
	"sync"
	"encoding/json"
)

type Event struct {
	Host string          `json:"host"`
	Service string       `json:"service"`
	State string         `json:"state"`
	Time string          `json:"time"`
	Description string   `json:"description"`
	Tags []string        `json:"tags"`
	Metric string        `json:"metric"`
}


type EventsList struct {
	sync.RWMutex
	data []Event
}

func NewEventsList() *EventsList {
	return &EventsList{
		data: make([]Event, 0),
	}
}

func (list *EventsList)Data() []Event {
	return list.data
}

func (list *EventsList)DataJSON() []byte {
	list.RLock()
	defer list.RUnlock()

	if bytes, err := json.Marshal(list.data); err == nil {
		return bytes
	} else {
		return []byte("[]")
	}
}

func (list *EventsList)Add(e Event) {
	list.Lock()
	list.data = append(list.data, e)
	list.Unlock()
}
