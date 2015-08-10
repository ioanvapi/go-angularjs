package main
import (
	"sync"
	"encoding/json"
	"strings"
	"log"
)

type Event struct {
	HostService
	State       string    `json:"state"`
	// keep time in unix format and it will be converted to locale by each client
	Time        int64     `json:"time"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Metric      float64   `json:"metric"`
	// when an event is acknowledged we store who and when did it and some message
	AckUser     string    `json:"ackUser"`
	AckMessage  string    `json:"ackMessage"`
	AckTime     int64     `json:"-"`
}


type EventsStore interface {
	AddActiveEvent(event Event) error
	AddAckEvent(event Event) error
	AllActiveEvents() (map[HostService]Event, error)
	AllAckEvents() (map[HostService]Event, error)
	DeleteActiveEvent(event Event) error
	DeleteAckEvent(event Event) error
}

type HostService struct {
	Host    string    `json:"host"`
	Service string    `json:"service"`
}

type EventsList struct {
	sync.RWMutex
	data map[HostService]Event
}

type DashboardEvents struct {
	ActiveEvents EventsList
	AckEvents EventsList
	db EventsStore
}

func (e *Event) isOK() bool {
	return strings.ToLower(strings.TrimSpace(e.State)) == "ok"
}

func (e *Event) copyAckFrom(ackEvent Event) {
	e.AckUser = ackEvent.AckUser
	e.AckMessage = ackEvent.AckMessage
	e.AckTime = ackEvent.AckTime
}

func (el *EventsList) copy() []Event {
	el.RLock()
	defer el.RUnlock()

	newList := make([]Event, 0)
	for _, event := range el.data {
		newList = append(newList, event)
	}
	return newList
}

func NewDashboardEvents(db EventsStore) (*DashboardEvents, error) {
	dashboardEvents := &DashboardEvents{db: db}

	// load events from data store for initial state
	if err := dashboardEvents.init(); err != nil {
		return nil, err
	}

	return dashboardEvents, nil
}

func (dash *DashboardEvents) ActiveEventsDataJSON() []byte {

	var activeEvents = struct {
		ActiveEvents  []Event    `json:"activeEvents"`
	}{dash.ActiveEvents.copy() }

	if bytes, err := json.Marshal(activeEvents); err == nil {
		return bytes
	} else {
		log.Println("error encoding in ActiveEventsDataJSON(): ", err)
		return []byte("[]")
	}
}

func (dash *DashboardEvents) AckEventsDataJSON() []byte {
	var ackEvents = struct {
		AckEvents  []Event    `json:"ackEvents"`
	}{dash.AckEvents.copy() }

	if bytes, err := json.Marshal(ackEvents); err == nil {
		return bytes
	} else {
		log.Println("error encoding in AckEventsDataJSON(): ", err)
		return []byte("[]")
	}
}

func (dash *DashboardEvents) DataJSON() []byte {
	activeEvents := dash.ActiveEvents.copy()
	ackEvents := dash.AckEvents.copy()

	var dashEvents = struct {
		ActiveEvents  []Event    `json:"activeEvents"`
		AckEvents     []Event    `json:"ackEvents"`
	}{
		ActiveEvents: activeEvents,
		AckEvents: ackEvents,
	}

	if bytes, err := json.Marshal(dashEvents); err == nil {
		return bytes
	} else {
		log.Println("error encoding in DataJSON(): ", err)
		return []byte("[]")
	}
}

// Transfer an active event to ack events after it's updated with ack info
// from argument
func (dash *DashboardEvents) MoveActive2AckEvents(ackEvent Event) {
	activeEvents := dash.ActiveEvents
	ackEvents := dash.AckEvents
	activeEvents.Lock()
	ackEvents.Lock()
	// save the actual active event
	ae := activeEvents.data[ackEvent.HostService]
	delete(activeEvents.data, ackEvent.HostService)

	// transfer ACK info to existing active event then push it to ack events
	ae.copyAckFrom(ackEvent)
	ackEvents.data[ackEvent.HostService] = ae

	activeEvents.Unlock()
	ackEvents.Unlock()
}

// analyze the event and decided if the active or ack list of events should be updated
// returns true if any list is updated
func (dash *DashboardEvents) Update(newEvent Event) bool {
	activeEvents := dash.ActiveEvents
	ackEvents := dash.AckEvents
	activeEvents.Lock()
	ackEvents.Lock()
	updated := false
	ackEvent, existsInAck := ackEvents.data[newEvent.HostService]
	if existsInAck { // then is should not exists in active
		// update state and time
		ackEvent.State = newEvent.State
		ackEvent.Time = newEvent.Time
		ackEvents.data[newEvent.HostService] = ackEvent
		updated = true
	} else {
		_, exists := activeEvents.data[newEvent.HostService]
		// add new event which is '!ok'
		// or update existing '!ok' event with a new '!ok' state
		if !newEvent.isOK() {
			activeEvents.data[newEvent.HostService] = newEvent
			updated = true
		}
		//remove existing '!ok' because a new 'ok' event came
		if newEvent.isOK() && exists {
			delete(activeEvents.data, newEvent.HostService)
			updated = true
		}
	}
	activeEvents.Unlock()
	ackEvents.Unlock()
	return updated
}

func (dash *DashboardEvents) init() error {
	// load active events from data store
	activeEvents, err := dash.db.AllActiveEvents()
	log.Printf("Loading %d active events from data store.\n", len(activeEvents))
	if err != nil {
		return err
	}
	dash.ActiveEvents = EventsList{data: activeEvents}

	// load ack events from data store
	ackEvents, err := dash.db.AllAckEvents()
	log.Printf("Loading %d ack events from data store.\n", len(activeEvents))
	if err != nil {
		return err
	}
	dash.AckEvents = EventsList{data: ackEvents}

	return nil
}