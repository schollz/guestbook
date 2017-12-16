package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/schollz/jsonstore"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

var ks *jsonstore.JSONStore

type Entry struct {
	Name    string
	Email   string
	Message string
	Date    time.Time
}

func init() {
	var err error
	ks, err = jsonstore.Open("guestbook.json.gz")
	if err != nil {
		ks = new(jsonstore.JSONStore)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "HI")
}

func jsonpHandler(w http.ResponseWriter, r *http.Request) {
	callbackName := r.URL.Query().Get("callback")
	message := r.URL.Query().Get("message")
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")
	if callbackName == "" {
		fmt.Fprintf(w, "Please give callback name in query string")
		return
	}
	if message != "" && name != "" {
		p := bluemonday.UGCPolicy()
		entry := Entry{
			Name:    p.Sanitize(name),
			Email:   p.Sanitize(email),
			Message: p.Sanitize(string(blackfriday.Run([]byte(message)))),
			Date:    time.Now(),
		}
		ks.Set(r.Header.Get("Referer")+":"+time.Now().String(), entry)
		go jsonstore.Save(ks, "guestbook.json.gz")
	}

	keys := ks.GetAll(regexp.MustCompile(r.Header.Get("Referer")))
	keyList := make([]string, len(keys))
	messages := make(map[string]Entry)
	i := 0
	for key := range keys {
		var entry Entry
		json.Unmarshal(keys[key], &entry)
		messages[key] = entry
		keyList[i] = key
		i++
	}
	sort.Strings(keyList)

	messageList := make([]Entry, len(keys))
	for i, key := range keyList {
		messageList[len(keys)-i-1] = messages[key]
	}
	b, err := json.Marshal(messageList)
	if err != nil {
		fmt.Fprintf(w, "json encode error")
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, "%s(%s);", callbackName, b)
}

func main() {
	http.HandleFunc("/jsonp", jsonpHandler)
	http.HandleFunc("/", handler)
	fmt.Println("Running at :8054")
	http.ListenAndServe(":8054", nil)
}
