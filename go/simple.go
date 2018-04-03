//usr/bin/env go run $0 $@; exit $?
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	customer         = "CUSTOMER"
	zone             = "ZONE"
	password         = "PASSWORD"
	port             = 22225
	username         = "lum-customer-%v-zone-%v" // "lum-customer-CUSTOMER-zone-ZONE"
	proxyTemplateURL = "http://%v-session-%v:%v@zproxy.luminati.io:%v"
)

func main() {
	if c := os.Getenv("CUSTOMER"); c != "" {
		customer = c
	}
	if z := os.Getenv("ZONE"); z != "" {
		zone = z
	}
	if p := os.Getenv("PASSWORD"); p != "" {
		password = p
	}
	username = fmt.Sprintf(username, customer, zone)
	if u := os.Getenv("USERNAME"); u != "" {
		username = u
	}
	rand.Seed(time.Now().UnixNano())
	sessionId := rand.Intn(10000)
	proxyUrlStr := fmt.Sprintf(proxyTemplateURL, username, sessionId, password, port)
	proxyUrl, err := url.Parse(proxyUrlStr)
	if err != nil {
		log.Fatalf("Error parsing proxy URL: %v", err)
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}
	req, err := http.NewRequest("GET", "http://lumtest.com/myip.json", nil)
	req.Header.Add("Accept", "text/plain")
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Fatalf("Cannot get data: %v", err)
	}
	defer resp.Body.Close()
	log.Printf("Response: %+v, err %v", resp, err)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("Got data: %s", body)
}
