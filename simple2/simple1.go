package main
import (
    "github.com/gorilla/mux"
    "net/http"
    "log"
    "os"
    "encoding/json"

)



func main() {
    r := mux.NewRouter()
    r.HandleFunc("/api/persons", GetPersons).Methods("GET")

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
    log.Println("I'm in handler ...")
//    fmt.Fprintln(w, `["Gigi", "Vasile", "Ana"]`)

    res := []string{"Gigi", "Vasile", "Ana"}
    err := json.NewEncoder(w).Encode(res)

    if err != nil {
        log.Println(err)
        http.Error(w, "oops", http.StatusInternalServerError)
    }

}