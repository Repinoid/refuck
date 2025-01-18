package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"gorono/internal/memos"
	"gorono/internal/middlas"
	"gorono/internal/models"

	"github.com/go-resty/resty/v2"
)

type MemStorage struct {
	gau    map[string]models.Gauge
	count  map[string]models.Counter
	mutter sync.RWMutex
}

// var memStor *MemStorage
var host = "localhost:8080"
var reportInterval = 10
var pollInterval = 2

func main() {
	if err := foa4Agent(); err != nil {
		log.Fatal("INTERVAL error ", err)
		return
	}
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	memStor := MemStorage{} //memStor := new(MemStorage)
	for {
		cunt := 0
		for i := 0; i < reportInterval/pollInterval; i++ {
			err := memos.GetMetrix(&memStor)
			if err != nil {
				log.Println(err, "getMetrix")
			} else {
				cunt++
			}
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}

		memStor.count["PollCount"] = counter(cunt)
		bunch := makeBunchOfMetrics(&memStor)
		log.Println(len(bunch))

		err := postBunch(bunch)
		if err != nil {
			log.Printf("AGENT postBunch ERROR %+v\n", err)
		}
	}
}

func postBunch(bunch []Metrics) error {
	marshalledBunch, err := json.Marshal(bunch)
	if err != nil {
		return err
	}
	compressedBunch, err := middlas.Pack2gzip(marshalledBunch)
	if err != nil {
		return err
	}
	httpc := resty.New() //
	httpc.SetBaseURL("http://" + host)

	httpc.SetRetryCount(3)
	httpc.SetRetryWaitTime(1 * time.Second)    // начальное время повтора
	httpc.SetRetryMaxWaitTime(9 * time.Second) // 1+3+5
	//tn := time.Now()                           // -------------
	httpc.SetRetryAfter(func(client *resty.Client, resp *resty.Response) (time.Duration, error) {
		rwt := client.RetryWaitTime
		//	fmt.Printf("waittime \t%+v\t time %+v  count %d\n", rwt, time.Since(tn), client.RetryCount) // -------
		client.SetRetryWaitTime(rwt + 2*time.Second)
		//	tn = time.Now() // ----------------
		return client.RetryWaitTime, nil
	})

	req := httpc.R().
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressedBunch).
		SetHeader("Accept-Encoding", "gzip")

	_, err = req.
		SetDoNotParseResponse(false).
		Post("/updates/")

		//	log.Printf("%+v\n", resp)

	return err
}
