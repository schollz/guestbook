package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
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
		var alreadyDid bool
		guestBookLimitString := ipAddress + ":" + r.Header.Get("Referer") + time.Now().Format("2006-01-02")
		ks.Get(guestBookLimitString, &alreadyDid)
		if !alreadyDid {
			ks.Set(guestBookLimitString, true)
			p := bluemonday.UGCPolicy()
			entry := Entry{
				Name:    p.Sanitize(name),
				Email:   p.Sanitize(email),
				Message: p.Sanitize(string(blackfriday.Run([]byte(message)))),
				Date:    time.Now(),
			}
			ks.Set(r.Header.Get("Referer")+":"+time.Now().String(), entry)
			go jsonstore.Save(ks, "guestbook.json.gz")
		} else {
			userMessage = "Sorry, only one entry per IP address is allowed. You can't sign a Guestbook twice!."
		}
	} else {
		userMessage = "Must include name and a message!"
	}

	keys := ks.GetAll(regexp.MustCompile(r.Header.Get("Referer") + ":"))
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
	http.HandleFunc("/jsonp", jsonpHandler)
	http.HandleFunc("/", handler)
	fmt.Println("Running at :8054")
	http.ListenAndServe(":8054", nil)
}

// GetClientIPHelper gets the client IP using a mixture of techniques.
// This is how it is with golang at the moment.
func GetClientIPHelper(req *http.Request) (ipResult string, errResult error) {

	// Try lots of ways :) Order is important.
	// Try Request Headers (X-Forwarder). Client could be behind a Proxy
	ip, err := getClientIPByHeaders(req)
	if err == nil {
		// log.Printf("debug: Found IP using Request Headers sniffing. ip: %v", ip)
		return ip, nil
	}

	// Try by Request
	ip, err = getClientIPByRequestRemoteAddr(req)
	if err == nil {
		// log.Printf("debug: Found IP using Request sniffing. ip: %v", ip)
		return ip, nil
	}

	//  Try Request Header ("Origin")
	url, err := url.Parse(req.Header.Get("Origin"))
	if err == nil {
		host := url.Host
		ip, _, err := net.SplitHostPort(host)
		if err == nil {
			// log.Printf("debug: Found IP using Header (Origin) sniffing. ip: %v", ip)
			return ip, nil
		}
	}

	err = errors.New("error: Could not find clients IP address")
	return "", err
}

// getClientIPByRequest tries to get directly from the Request.
// https://blog.golang.org/context/userip/userip.go
func getClientIPByRequestRemoteAddr(req *http.Request) (ip string, err error) {

	// Try via request
	ip, _, err = net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		// log.Printf("debug: Getting req.RemoteAddr %v", err)
		return "", err
	} else {
		// log.Printf("debug: With req.RemoteAddr found IP:%v; Port: %v", ip, port)
	}

	userIP := net.ParseIP(ip)
	if userIP == nil {
		message := fmt.Sprintf("debug: Parsing IP from Request.RemoteAddr got nothing.")
		// log.Printf(message)
		return "", fmt.Errorf(message)

	}
	// log.Printf("debug: Found IP: %v", userIP)
	return userIP.String(), nil

}

// getClientIPByHeaders tries to get directly from the Request Headers.
// This is only way when the client is behind a Proxy.
func getClientIPByHeaders(req *http.Request) (ip string, err error) {

	// Client could be behid a Proxy, so Try Request Headers (X-Forwarder)
	ipSlice := []string{}

	ipSlice = append(ipSlice, req.Header.Get("X-Forwarded-For"))
	ipSlice = append(ipSlice, req.Header.Get("x-forwarded-for"))
	ipSlice = append(ipSlice, req.Header.Get("X-FORWARDED-FOR"))

	for _, v := range ipSlice {
		// log.Printf("debug: client request header check gives ip: %v", v)
		if v != "" {
			return v, nil
		}
	}
	err = errors.New("error: Could not find clients IP address from the Request Headers")
	return "", err

}
