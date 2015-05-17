package main
import (
    "net/http"
    "log"
    "fmt"
    "os"
)

/*
Message from /api/groups looks like
{
    "Groups" :[{
        "Name":"Gigi Marga",
        "URL":"http://ceva.com/mama",
        "Members":194,
        "City":"San Mateo",
        "Country":"US"
    }, {
        "Name":"Stefan Vlahuta",
        "URL":"http://ceva.com/tata",
        "Members":1393,
        "City":"San Francisco",
        "Country":"US"
    }],
    "Errors" :[
        "something bad happened."
    ]
}
*/

const response = `
{
    "Groups" :[{
        "Name":"Gigi Marga",
        "URL":"http://ceva.com/mama",
        "Members":194,
        "City":"San Mateo",
        "Country":"US"
    }, {
        "Name":"Stefan Vlahuta",
        "URL":"http://ceva.com/tata",
        "Members":1393,
        "City":"San Francisco",
        "Country":"US"
    }],
    "Errors" :[
        "something bad happened."
    ]
}`



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


// check - http://localhost:8080/api/groups  (json is formated in browser)
// check - http://localhost:8080
func GetGroups(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, response)
}