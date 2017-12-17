package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
)

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

func LocationFromIP(ip string) (location string) {
	type ResultJSON struct {
		IP          string  `json:"ip"`
		CountryCode string  `json:"country_code"`
		CountryName string  `json:"country_name"`
		RegionCode  string  `json:"region_code"`
		RegionName  string  `json:"region_name"`
		City        string  `json:"city"`
		ZipCode     string  `json:"zip_code"`
		TimeZone    string  `json:"time_zone"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		MetroCode   int     `json:"metro_code"`
	}
	resp, err := http.Get("http://geoip.makemydrive.fun" + "/json/" + ip)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var result ResultJSON
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return
	}
	location = fmt.Sprintf("%s, %s, %s", result.City, result.RegionName, result.CountryName)
	if len(location) < 5 || err != nil {
		location = ""
	}
	return
}
