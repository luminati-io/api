//usr/bin/env go run $0 $@; exit $?
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	username         = "lum-customer-CUSTOMER-zone-ZONE"
	password         = "PASSWORD"
	port             = 22225
	proxyTemplateURL = "http://%v-session-%v:%v@zproxy.luminati.io:%v"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	session_id := rand.Intn(10000)
	proxyUrlStr := fmt.Sprintf(proxyTemplateURL, username, session_id, password, port)
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
	log.Printf("Resp: %+v, err %v", resp, err)
	body, err := ioutil.ReadAll(resp.Body)
	log.Printf("Got data: %s", body)
}
