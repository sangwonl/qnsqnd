package main

import (
	"encoding/json"
	"fmt"
	"github.com/sangwonl/qnsqnd/pkg/core"
	"gopkg.in/resty.v1"
	"log"
	"flag"
	"net/http"
	"sync"
	"time"
)

func main() {
	concurrency := flag.Int("concurrency", 10, "Concurent connection to the server")
	duration := flag.Int("duration", 10, "Duration of subscription test (in seconds)")
	host := flag.String("host", "localhost:8080", "Target host")
	topic := flag.String("topic", "", "Topic to subscribe")
	filter := flag.String("filter", "", "Filter expression")

	flag.Parse()

	log.Printf("Current concurrency: %d\n", *concurrency)
	log.Printf("Duration: %d\n", *duration)
	log.Printf("Target host: %s\n", *host)
	log.Printf("Topic to subscribe: %s\n", *topic)
	log.Printf("Expression to filter message: %s\n", *filter)

	r := resty.New().SetTransport(&http.Transport{
		MaxIdleConns: *concurrency,
		MaxIdleConnsPerHost: *concurrency,
	})

	var wg sync.WaitGroup

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			url := fmt.Sprintf("http://%s/subscribe", *host)
			res, err := r.R().
				SetHeader("Content-Type", "application/json").
				SetBody(core.Subscription{Topic: core.Topic(*topic), Filter: *filter}).
				Post(url)

			if err != nil {
				log.Printf(err.Error())
				return
			}

			data := make(map[string]interface{})
			_ = json.Unmarshal(res.Body(), &data)

			subId := data["subscriptionId"]
			timer := time.NewTimer(time.Duration(*duration) * time.Second)
			go func() {
				<-timer.C
				url = fmt.Sprintf("http://%s/subscribe/%s/cancel", *host, subId)
				log.Printf("SSE Subscription Url: %s\n", url)
				r.R().Post(url)
			}()

			url = fmt.Sprintf("http://%s/subscribe/%s/sse", *host, subId)
			log.Printf("SSE Subscription Url: %s\n", url)
			r.R().Get(url)
		}()
	}

	wg.Wait()
}

// refs..
// https://github.com/jiaz/simpleloadclient/blob/master/main.go