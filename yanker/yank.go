package yanker

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type Yank struct {
	url string
	ccn int
}

type job struct {
	url   string
	index string
}

func NewYankManager(url string, options ...Options) *Yank {
	ccn := 4 //default
	if len(options) > 0 {
		if c := options[0].ConcurrentConnections; c > 0 {
			ccn = c
		}
	}

	return &Yank{url: url, ccn: ccn}

}

func (y *Yank) StartDownload() ([]byte, error) {
	length, support, err := checkRangeRequestSupport(y.url)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	log.Println("Content-Length: ", length)
	log.Println("Ranged Support: ", support)

	filePrefix, _ := generateFileName()

	contentLength, _ := strconv.Atoi(length)

	//chunks := splitFileIntoChunks(414111315, y.ccn)
	chunks := splitFileIntoChunks(contentLength, y.ccn)

	wg := sync.WaitGroup{}
	stopChan := make(chan struct{})

	wg.Add(y.ccn)

	for index, rng := range chunks {
		go func(i int, r string) {
			defer wg.Done()
			download(filePrefix, strconv.Itoa(i), r, y.url)
		}(index, rng)
	}

	go startSpeedMonitor(filePrefix, y.ccn, stopChan)

	wg.Wait()
	<-stopChan

	fmt.Println("Finished Doing Nonsense")
	return nil, nil
}

func checkRangeRequestSupport(url string) (string, bool, error) {
	req, err := http.NewRequest("HEAD", url, nil)

	if err != nil {
		return "", false, err
	}

	req.Header.Set("Range", "bytes=0-100")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", false, err
	}

	if resp.StatusCode != 200 {
		req.Header.Set("Range", "")
		req.Method = "GET"

		resp, err = http.DefaultClient.Do(req)

		if err != nil {
			return "", false, err
		}

		if resp.StatusCode != 200 {
			return "", false, errors.New(fmt.Sprintf("non 200 status code received - %d\n", resp.StatusCode))
		}
	}

	rangeSupport := resp.Header.Get("Accept-Ranges")
	contentLength := resp.Header.Get("Content-Length")

	log.Println(rangeSupport)
	log.Println(contentLength)
	if rangeSupport == "bytes" {
		return contentLength, true, nil
	}
	return contentLength, false, nil
}
