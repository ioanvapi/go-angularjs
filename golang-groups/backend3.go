package main
import (
    "net/http"
    "log"
    "fmt"
    "os"
    "encoding/json"
    "strings"
)


type Group struct {
    Name string     `json:"name"`
    URL string      `json:"url"`
    Members int     `json:"members"`
    City string     `json:"city"`
    Country string  `json:"country"`
}

// golang-rust is an invalid group name and we will receive an error message for it from meetup.com
var ids = []string{"golangsf", "golangsv", "golang-rust"}

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

//

// check - http://localhost:8080/api/groups  (json is NOT formated in browser)
// check - http://localhost:8080
func GetGroups(w http.ResponseWriter, r *http.Request) {
    res := struct {
        Groups []Group      `json:"groups"`
        Errors []string     `json:"errors"`
    }{}

    for _, id := range ids {
        g, err := fetch(id)
        if err != nil {
            res.Errors = append(res.Errors, err.Error())
            continue
        }
        res.Groups = append(res.Groups, *g)
    }

    // can use err := json.NewEncoder(w).Encode(res)
    data, err := json.Marshal(res)
    if err != nil {
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return

    }
    fmt.Fprintln(w, string(data))
}


// in case of querying something invalid we get this type of response
/*
{
    "errors": [
        {
            "code": "group_error",
            "message": "Invalid group urlname"
        }
    ]
}
*/
/*
the proper response json is like - and more
{
    "id": 4829842,
    "name": "GoSV",
    "link": "http://www.meetup.com/GolangSV/",
    "urlname": "GolangSV",
    "description": "An informal meetup to talk about or develop with the Go Programming Language (golang).",
    "created": 1347112621000,
    "city": "San Mateo",
    "country": "US",
    "state": "CA",
    "join_mode": "open",
    "visibility": "public",
    "lat": 37.540000915527344,
    "lon": -122.30000305175781,
}
*/
// we want to be able to decode both situations (error/success) in the same struct type
func fetch(id string) (*Group, error) {
    const (
    // I have generated this key
    // https://api.meetup.com/foo?sign=true&key=66b796a31a197318546c3140e793
        apiKey = "66b796a31a197318546c3140e793"
        urlTemplate = "https://api.meetup.com/%s?sign=true&key=%s"
    )
    url := fmt.Sprintf(urlTemplate, id, apiKey)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }

    //    if resp.StatusCode != 200 {
    //        return nil, fmt.Errorf("error fetching data from meetups (got status code '%d')", resp.StatusCode)
    //    }

    var g struct {
        Name    string  `json:"name"`
        Link    string  `json:"link"`
        City    string  `json:"city"`
        Country string  `json:"country"`
        Members int     `json:"members"`
        Errors []struct{
            Code string     `json:"code"`
            Message string  `json:"message"`
        } `json:"errors"`
    }

    dec := json.NewDecoder(resp.Body)
    err = dec.Decode(&g)
    if err != nil {
        return nil, fmt.Errorf("decode: %v", err)
    }

    if len(g.Errors) > 0 {
        var errs []string
        for _, e := range g.Errors {
            errs = append(errs, e.Message)
        }
        return nil, fmt.Errorf(strings.Join(errs, "/n"))
    }

    return &Group {
        Name: g.Name,
        URL: g.Link,
        Members: g.Members,
        City: g.City,
        Country:g.Country,
    }, nil
}