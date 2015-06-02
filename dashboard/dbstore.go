package main
import (
	"github.com/boltdb/bolt"
	"fmt"
	"encoding/json"
)


var (
	dbActiveEvents = []byte("active-events")
	dbAckEvents = []byte("ack-events")
	dbHistoryEvents = []byte("history-events")
)

type EventsStore struct {
	conn *bolt.DB
	path string
}


func NewEventsStore(path string) (*EventsStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	store := &EventsStore{
		conn: db,
		path: path,
	}

	if err = store.init(); err != nil {
		store.Close()
		return nil, err
	}

	return store, nil
}

func (db *EventsStore) Close() {
	db.conn.Close()
}

func (db *EventsStore) init() error {
	err := db.conn.Update(func (tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(dbActiveEvents); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(dbAckEvents); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(dbHistoryEvents); err != nil {
			return err
		}
		return nil
	})
	return err
}


func (db *EventsStore) AddActiveEvent(event Event) error {
	return db.addEvent(event, dbActiveEvents)
}

func (db *EventsStore) AddAckEvent(event Event) error {
	return db.addEvent(event, dbAckEvents)
}

func (db *EventsStore) AllActiveEvents() (map[HostService]Event, error) {
	return db.allEvents(dbActiveEvents)
}

func (db *EventsStore) AllAckEvents() (map[HostService]Event, error) {
	return db.allEvents(dbAckEvents)
}

func (db *EventsStore) DeleteActiveEvent(event Event) error {
	return db.deleteEvent(event, dbActiveEvents)
}

func (db *EventsStore) DeleteAckEvent(event Event) error {
	return db.deleteEvent(event, dbAckEvents)
}


func (db *EventsStore) addEvent(event Event, bucketName []byte) error {
	err := db.conn.Update(func (tx *bolt.Tx) error {
		eventsBucket := tx.Bucket(bucketName)

		dbValue, err := dbValue(event)
		if err != nil {
			return err
		}

		if err := eventsBucket.Put(dbKey(event.HostService), dbValue); err != nil {
			return err
		}
		return nil
	})
	return err
}

func (db *EventsStore) deleteEvent(event Event, bucketName []byte) error {
	err := db.conn.Update(func (tx *bolt.Tx) error {
		return tx.Bucket(bucketName).Delete(dbKey(event.HostService))
	})
	return err
}

func (db *EventsStore) allEvents(bucketName []byte) (map[HostService]Event, error) {
	events := make(map[HostService]Event)
	err := db.conn.View(func (tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		b.ForEach(func (k, v []byte) error {
			dbEvent, err := dbEvent(v)
			if err != nil {
				return err
			}
			events[(*dbEvent).HostService] = *dbEvent
			return nil
		})
		return nil
	})
	return events, err
}


func dbKey(hs HostService) []byte {
	return []byte(fmt.Sprintf("%s:%s", hs.Host, hs.Service))
}

func dbValue(event Event) ([]byte, error) {
	return json.Marshal(event)
}

func dbEvent(data []byte)(*Event, error) {
	event := new(Event)
	if err := json.Unmarshal(data, event); err != nil {
		return nil, err
	}
	return event, nil
}