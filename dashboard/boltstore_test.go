package main
import (
	"testing"
	"os"
	"github.com/boltdb/bolt"
	"fmt"
	"time"
	"encoding/json"
	"log"
	"strconv"
	"bytes"
)


func TestNewStoreCreated(t *testing.T) {
	fileName := "t1.db"
	db, err := NewBoltStore(fileName)
	if err != nil {
		t.Fatalf("error creating NewEventsStore() to '%s'", fileName)
	}
	defer func() {
		db.Close()
		os.Remove(fileName)
	}()


	err = db.conn.View(func (tx *bolt.Tx) error {
		if ae := tx.Bucket(dbActiveEvents); ae == nil {
			return fmt.Errorf("bucket %s not create in init()", string(dbActiveEvents))
		}
		if ae := tx.Bucket(dbAckEvents); ae == nil {
			return fmt.Errorf("bucket %s not create in init()", string(dbAckEvents))
		}
		if ae := tx.Bucket(dbHistoryEvents); ae == nil {
			return fmt.Errorf("bucket %s not create in init()", string(dbHistoryEvents))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("error %v", err)
	}
}


func TestWhenAddActiveEvent_ThenItIsStored(t *testing.T) {
	fileName := "t2.db"
	db, err := NewBoltStore(fileName)
	if err != nil {
		t.Fatalf("error creating NewEventsStore() to '%s'", fileName)
	}
	defer func() {
		db.Close()
		os.Remove(fileName)
	}()

	ae := Event{
		HostService: HostService{"Host1", "Service1"},
		State: "ok",
		Time: time.Now().Unix(),
		Description: "TestWhenAddActiveEvent_ThenItIsStored",
	}
	err = db.AddActiveEvent(ae)
	if err != nil {
		t.Fatalf("error in AddActiveEvent(): %v", err)
	}

	event := getEventFromStore(db, ae.HostService, dbActiveEvents)
	if event == nil {
		t.Fatalf("Fail getting %v from store", ae.HostService)
	}
	if ae.Host != event.Host || ae.Service != event.Service || ae.Description != event.Description {
		log.Println(event)
		t.Fatalf("expected %v but got %v", ae, event)
	}
}


func TestWhenAddSomeActiveEvents_ThenGetAllActiveEvents(t *testing.T) {
	fileName := "t3.db"
	db, err := NewBoltStore(fileName)
	if err != nil {
		t.Fatalf("error creating NewEventsStore() to '%s'", fileName)
	}
	defer func() {
		db.Close()
		os.Remove(fileName)
	}()

	hosts := make([]string, 0)
	services := make([]string, 0)
	for i := 1; i < 4; i++ {
		si := strconv.Itoa(i)
		host, service := "Host" + si, "Service" + si
		ae := Event{
			HostService: HostService{host, service},
			State: "ok",
			Time: time.Now().Unix(),
			Description: "TestWhenAddActiveEvent_ThenItIsStored " + si,
		}
		addEventToStore(db, ae, dbActiveEvents)

		hosts = append(hosts, host)
		services = append(services, service)
	}

	all, err := db.AllActiveEvents()
	if err != nil {
		t.Fatalf("error when executing AllActiveEvents()", err)
	}

	if len(all) != 3 {
		t.Fatalf("expect %d events but got %d events", 3, len(all))
	}

	for _, event := range all {
		if !isIn(event.Host, hosts) {
			t.Fatalf("Invalid host detected '%s'", event.Host)
		}
		if !isIn(event.Service, services){
			t.Fatalf("Invalid service detected '%s'", event.Service)
		}
	}
}

func TestWhenDeleteAnActiveEvent_ThenItIsNoLongerInStore(t *testing.T) {
	fileName := "t4.db"
	db, err := NewBoltStore(fileName)
	if err != nil {
		t.Fatalf("error creating NewEventsStore() to '%s'", fileName)
	}
	defer func() {
		db.Close()
		os.Remove(fileName)
	}()

	hosts := make([]string, 0)
	services := make([]string, 0)
	for i := 1; i < 4; i++ {
		si := strconv.Itoa(i)
		host, service := "Host" + si, "Service" + si
		ae := Event{
			HostService: HostService{host, service},
			State: "ok",
			Time: time.Now().Unix(),
			Description: "TestWhenAddActiveEvent_ThenItIsStored " + si,
		}
		addEventToStore(db, ae, dbActiveEvents)

		hosts = append(hosts, host)
		services = append(services, service)
	}

	err = db.DeleteActiveEvent(Event{HostService : HostService{"Host2", "Service2"}})
	if err != nil {
		t.Fatal("error deleting event from active events")
	}

	err = db.conn.View(func (tx *bolt.Tx) error {
		counter := 0
		found := false
		key := dbKey(HostService{"Host2", "service2"})

		if err := tx.Bucket(dbActiveEvents).ForEach(func (k, v []byte) error {
			if bytes.Equal(key, k) {
				found = true
			}
			counter++
			return nil
		}); err != nil {
			t.Fatal("error ForEach() in active events")
		}
		if counter != 2 {
			t.Fatalf("expected %d events in store but found %d", 2, counter)
		}
		if found {
			t.Fatalf("expect not to found event with key %v ", key)
		}
		return nil
	})
	if err != nil {
		t.Fatal("error checking event deleted from active events")
	}
}

func addEventToStore(db *BoltStore, event Event, bucketName []byte) {
	err := db.conn.Update(func (tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		dbValue, err := dbValue(event)
		if err != nil {
			return fmt.Errorf("errro getting dbValue %v", err)
		}
		bucket.Put(dbKey(event.HostService), dbValue)
		return nil
	})
	if err != nil {
		log.Println("Error when updating store: ", err)
	}
}


func getEventFromStore(db *BoltStore, hs HostService, bucketName []byte) *Event {
	var event Event
	err := db.conn.View(func (tx *bolt.Tx) error {
		data := tx.Bucket(bucketName).Get(dbKey(hs))
		if data == nil {
			return fmt.Errorf("error getting from db %s", string(bucketName))
		}
		if err := json.Unmarshal(data, &event); err != nil {
			return fmt.Errorf("error unmarshal data %v", data)
		}
		return nil
	})
	if err != nil {
		log.Println("Error gettting event %v", err)
	}
	return &event
}

func isIn(s string, list []string) bool {
	for _, v := range list {
		if s == v {
			return true
		}
	}
	return false
}