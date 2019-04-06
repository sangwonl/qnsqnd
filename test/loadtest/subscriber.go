package main

import (
	"encoding/json"
	"fmt"
	"github.com/sangwonl/qnsqnd/pkg/core"
	"gopkg.in/resty.v1"
	"log"
	"flag"
	"time"
)

func main() {
	concurrency := flag.Int("concurrency", 10, "Concurent connection to the server")
	duration := flag.Int("duration", 10, "Duration of subscription test (in seconds)")
	host := flag.String("host", "localhost:8080", "Target host")
	topic := flag.String("topic", "", "Topic to subscribe")

	flag.Parse()

	log.Printf("Current concurrency: %d\n", *concurrency)
	log.Printf("Duration: %d\n", *duration)
	log.Printf("Target host: %s\n", *host)
	log.Printf("Subscribing topic: %s\n", *topic)

	fin := make(chan bool)

	for i := 0; i < *concurrency; i++ {
		go func() {
			url := fmt.Sprintf("http://%s/subscribe", *host)
			res, _ := resty.R().
				SetHeader("Content-Type", "application/json").
				SetBody(core.Subscription{Topic: core.Topic(*topic)}).
				Post(url)

			data := make(map[string]interface{})
			_ = json.Unmarshal(res.Body(), &data)

			subId := data["subscriptionId"]
			timer := time.NewTimer(time.Duration(*duration) * time.Second)
			go func() {
				<-timer.C
				url = fmt.Sprintf("http://%s/subscribe/%s/cancel", *host, subId)
				log.Printf("SSE Subscription Url: %s\n", url)
				resty.R().Post(url)
			}()

			url = fmt.Sprintf("http://%s/subscribe/%s/sse", *host, subId)
			log.Printf("SSE Subscription Url: %s\n", url)
			resty.R().Get(url)
		}()
	}

	<- fin
}

// refs..
// https://github.com/jiaz/simpleloadclient/blob/master/main.go