package main

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"time"

	"gorono/internal/memos"
	"gorono/internal/middlas"
	"gorono/internal/models"
	"gorono/internal/privacy"

	"github.com/go-resty/resty/v2"
)

var host = "localhost:8080"
var reportInterval = 10
var pollInterval = 2

func main() {
	if err := initAgent(); err != nil {
		log.Fatal("INTERVALS error ", err)
		return
	}
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	memStorage := []models.Metrics{}
	for {
		cunt := int64(0)
		for i := 0; i < reportInterval/pollInterval; i++ {
			memStorage = memos.GetMetrixFromOS()
			cunt++
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
		for ind, metr := range memStorage {
			if metr.ID == "PollCount" && metr.MType == "counter" {
				memStorage[ind].Delta = &cunt // в сам memStorage, metr - копия
				break
			}
		}
		err := postBunch(memStorage)
		if err != nil {
			log.Printf("AGENT postBunch ERROR %+v\n", err)
		}
	}
}

func postBunch(bunch []models.Metrics) error {
	marshalledBunch, err := json.Marshal(bunch)
	if err != nil {
		return err
	}

	keyB, _ := privacy.RandBytes(32)

	coded, err := privacy.EncryptB2B(marshalledBunch, keyB)
	if err != nil {
		return err
	}
	codedWkey := append(keyB, coded...)
	ha := privacy.MakeHash(nil, codedWkey, keyB)
	haHex := hex.EncodeToString(ha)

	compressedBunch, err := middlas.Pack2gzip(codedWkey)
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
		SetHeader("Content-Encoding", "gzip"). // сжаtо
		SetBody(compressedBunch).
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("HashSHA256", haHex)

	_, err = req.
		SetDoNotParseResponse(false).
		Post("/updates/") // slash on the tile

		//	log.Printf("%+v\n", resp)

	return err
}
