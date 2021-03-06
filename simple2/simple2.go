package main
import (
    "github.com/gorilla/mux"
    "net/http"
    "log"
    "os"
    "encoding/json"

)


var persons []string = []string{"Gigi", "Vasile", "Ana"};

func main() {
    r := mux.NewRouter()
    r.HandleFunc("/api/persons", GetPersons).Methods("GET")
    r.HandleFunc("/api/person", AddPerson).Methods("POST")

    http.Handle("/api/", r)
    wd, _ := os.Getwd()
    http.Handle("/", http.FileServer(http.Dir(wd + "/web")))


    log.Println("Starting server ...", wd)

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}


func GetPersons(w http.ResponseWriter, r *http.Request) {
    log.Println("I'm in handler 'GetPersons'...")

    err := json.NewEncoder(w).Encode(persons)

    if err != nil {
        log.Println(err)
        http.Error(w, "oops", http.StatusInternalServerError)
    }
}

func AddPerson(w http.ResponseWriter, r *http.Request) {
    log.Println("I'm in handler 'AddPerson'...")
    req := struct { Name string }{}

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Println("Adding new person ...", req.Name)
    persons = append(persons, req.Name)
}