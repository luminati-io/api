//usr/bin/env go run $0 $@; exit $?
//
// Multithreaded parallel version of proxychecker
//
// To run set env variables:
//   CUSTOMER - your customer name
//   ZONE - zone id
//   PASSWORD - your zone password

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

func getProxyUrl(task Task) (*url.URL, error) {
	var (
		customer         = getEnvOrDefault("CUSTOMER", "customer")
		zone             = getEnvOrDefault("ZONE", "zone")
		password         = getEnvOrDefault("PASSWORD", "password")
		port             = 22225
		username         = "lum-customer-%v-zone-%v" // "lum-customer-CUSTOMER-zone-ZONE"
		proxyTemplateURL = "http://%v-session-%v:%v@zproxy.luminati.io:%v"
	)
	username = fmt.Sprintf(username, customer, zone)
	proxyStringUrl := fmt.Sprintf(proxyTemplateURL, username, task.session, password, port)
	proxyUrl, err := url.Parse(proxyStringUrl)
	if err != nil {
		return nil, err
	}
	return proxyUrl, nil
}

func main() {
	var startAt = time.Now()
	// maximal number of parallel workers
	var maxInParallel = 10

	var taskChannel = make(ChanTask)
	var resultChannel = make(ChanResult)
	var quitChanel = make(chan struct{})

	// start result reader and printer to stdout
	go resultPrinter(resultChannel, quitChanel)

	// start page processor
	go startProcessors(taskChannel, resultChannel, maxInParallel)

	for _, taskInfo := range taskInfos {
		taskChannel <- Task{TaskInfo: taskInfo, startAt: time.Now()}
	}
	close(taskChannel)
	<-quitChanel
	log.Printf("Total time: %v", time.Since(startAt))
}

type TaskInfo struct {
	session string
	href    string
}
type Task struct {
	TaskInfo
	startAt, endAt time.Time
}
type ChanTask chan Task

type Result struct {
	session string
	startAt time.Time
	body    []byte
}
type ChanResult chan Result

func startProcessors(inCh ChanTask, outCh ChanResult, number int) {
	var wg sync.WaitGroup
	for i := 0; i < number; i++ {
		go processor(inCh, outCh, &wg)
		wg.Add(1)
	}
	wg.Wait()
	close(outCh)
}

func processor(inCh ChanTask, outCh ChanResult, wg *sync.WaitGroup) {
	for task := range inCh {
		proxyUrl, err := getProxyUrl(task)
		if err != nil {
			log.Printf("Error getting proxyURL: %v", err)
			continue // ignore this task
		}
		result, err := processPage(task, proxyUrl)
		if err != nil {
			log.Printf("Error processing page url %v: %v", task.href, err)
			continue // ignore this task
		}
		outCh <- *result
	}
	wg.Done()
}

func processPage(task Task, proxyUrl *url.URL) (result *Result, err error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
	}
	request, err := http.NewRequest("GET", task.href, nil)
	if err != nil {
		log.Printf("Error forming request: %v", err)
		return nil, err
	}
	request.Header.Add("Accept", "text/plain")
	resp, err := httpClient.Do(request)
	if err != nil {
		log.Printf("Cannot get data: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Result{session: task.session, startAt: task.startAt, body: body}, nil
}

// prints results,
// to stop printing close incoming channel inCh
// before exit, closes quitChat to inform about it
func resultPrinter(inCh ChanResult, quitChan chan struct{}) {
	var totalDuration time.Duration
	var i = 0
	for result := range inCh {
		var dur = time.Since(result.startAt)
		log.Printf("%d\tStartTime: %v\tDuration: %v\tSession: %- 8s\tBody: %s",
			i, result.startAt, dur, result.session, result.body)
		i++
		totalDuration += dur
	}
	close(quitChan)
	log.Printf("Total duration: %v", totalDuration)
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

var (
	taskInfos = []TaskInfo{
		{"12313", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"11", "http://lumtest.com/myip.json"},
		{"2323", "http://lumtest.com/myip.json"},
		{"rer", "http://lumtest.com/myip.json"},
		{"sdf", "http://lumtest.com/myip.json"},
		{"44", "http://lumtest.com/myip.json"},
		{"343", "http://lumtest.com/myip.json"},
		{"gegd", "http://lumtest.com/myip.json"},
		{"gegd", "http://lumtest.com/myip.json"},
		{"gegd", "http://lumtest.com/myip.json"},
		{"gegd", "http://lumtest.com/myip.json"},
		{"ett4", "http://lumtest.com/myip.json"},
		{"3r3r3r", "http://lumtest.com/myip.json"},
		{"34r3", "http://lumtest.com/myip.json"},
		{"fffd", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"sdff", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"sfsf", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"sfs", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"gegd", "http://lumtest.com/myip.json"},
		{"12313", "http://lumtest.com/myip.json"},
		{"00000", "http://lumtest.com/myip.json"},
	}
)
