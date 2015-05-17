package main
import (
    "net/http"
    "log"
    "fmt"
    "os"
    "encoding/json"
)

// user lower case names to generate lowercase names in json
type Group struct {
    Name string     `json:"name"`
    URL string      `json:"url"`
    Members int     `json:"members"`
    City string     `json:"city"`
    Country string  `json:"country"`
}

var groups = []Group{
    {"Ana Blandiana", "http://ceva.com/mama", 194, "San Mateo", "US"},
    {"Gigi Becali", "http://ceva.com/tata", 1393, "San Francisco", "US"},
}

func init() {
    http.HandleFunc("/api/groups", GetGroups)
}

func main() {

    wd, _ := os.Getwd()
    http.Handle("/", http.FileServer(http.Dir(wd + "/web")))

    fmt.Println("Start ListenAndServe()")
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}



// check - http://localhost:8080/api/groups  (json is NOT formated in browser)
// check - http://localhost:8080
func GetGroups(w http.ResponseWriter, r *http.Request) {
    res := struct{
        Groups []Group      `json:"groups"`
        Errors []string     `json:"errors"`
    }{
        groups,
        []string{"bad things happened"},
    }

    // can use err := json.NewEncoder(w).Encode(res)
    data, err := json.Marshal(res)
    if err != nil {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return

    }
    fmt.Fprintln(w, string(data))
}
