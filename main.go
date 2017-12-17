package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	Name       string
	Location   string
	Email      string
	Message    string
	DateString string
	Date       time.Time
}

type Flags struct {
	OneEntryPerPersonPerDay bool
	Port                    string
}

var flags Flags

func init() {
	var err error
	ks, err = jsonstore.Open("guestbook.json.gz")
	if err != nil {
		ks = new(jsonstore.JSONStore)
	}
	fmt.Println(LocationFromIP("198.199.67.130"))
}

func jsonpHandler(w http.ResponseWriter, r *http.Request) {
	var userMessage string
	callbackName := r.URL.Query().Get("callback")
	message := r.URL.Query().Get("message")
	name := r.URL.Query().Get("name")
	email := r.URL.Query().Get("email")
	ipAddress, err := GetClientIPHelper(r)
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	if callbackName == "" {
		fmt.Fprintf(w, "Please give callback name in query string")
		return
	}
	if message == "" && name == "" {
	} else if message != "" && name != "" {
		alreadyDid := false
		guestBookLimitString := ipAddress + ":" + r.Header.Get("Referer") + time.Now().Format("2006-01-02")
		if flags.OneEntryPerPersonPerDay {
			errGet := ks.Get(guestBookLimitString, &alreadyDid)
			if errGet != nil {
				alreadyDid = false
			}
		}
		fmt.Println(alreadyDid)
		if !alreadyDid {
			ks.Set(guestBookLimitString, true)
			p := bluemonday.UGCPolicy()
			entry := Entry{
				Name:     p.Sanitize(name),
				Email:    p.Sanitize(email),
				Message:  p.Sanitize(string(blackfriday.Run([]byte(message)))),
				Location: LocationFromIP(ipAddress),
				Date:     time.Now(),
			}
			ks.Set(r.Header.Get("Referer")+":"+time.Now().String(), entry)
			go jsonstore.Save(ks, "guestbook.json.gz")
		} else {
			userMessage = "Sorry, you can't sign a Guestbook more than once per day!"
		}
	} else {
		userMessage = "Please include name and a message."
	}

	keys := ks.GetAll(regexp.MustCompile(r.Header.Get("Referer") + ":"))
	keyList := make([]string, len(keys))
	messages := make(map[string]Entry)
	i := 0
	for key := range keys {
		var entry Entry
		json.Unmarshal(keys[key], &entry)
		entry.DateString = entry.Date.Format("January 2, 2006")
		messages[key] = entry
		keyList[i] = key
		i++
	}
	sort.Strings(keyList)

	messageList := make([]Entry, len(keys))
	for i, key := range keyList {
		messageList[len(keys)-i-1] = messages[key]
	}

	type Payload struct {
		Entries []Entry
		Message string
	}
	payload := Payload{
		Entries: messageList,
		Message: userMessage,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(w, "json encode error")
		return
	}

	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprintf(w, "%s(%s);", callbackName, b)
}

func main() {
	flags.OneEntryPerPersonPerDay = false
	http.HandleFunc("/jsonp", jsonpHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		index, _ := ioutil.ReadFile("index.html")
		fmt.Fprintf(w, string(index))
	})
	http.HandleFunc("/guestbook.css", func(w http.ResponseWriter, r *http.Request) {
		index, _ := ioutil.ReadFile("guestbook.css")
		fmt.Fprintf(w, string(index))
	})
	http.HandleFunc("/guestbook.js", func(w http.ResponseWriter, r *http.Request) {
		index, _ := ioutil.ReadFile("guestbook.js")
		fmt.Fprintf(w, string(index))
	})
	fmt.Println("Running at :8054")
	http.ListenAndServe(":8054", nil)
}
