package yanker

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type Yank struct {
	url      string
	ccn      int
	filename string
}

type job struct {
	url   string
	index string
}

func NewYankManager(url string, options ...Options) *Yank {
	ccn := 4 //default
	var filename string
	if len(options) > 0 {
		if c := options[0].ConcurrentConnections; c > 0 {
			ccn = c
		}
		filename = options[0].Filename
	}

	return &Yank{url: url, ccn: ccn, filename: filename}

}

func parseFileName(uri string) string {
	name := strings.Split(uri, "/")

	fileName, _ := url.QueryUnescape(name[len(name)-1])
	return fileName
}

func (y *Yank) StartDownload() ([]byte, error) {

	if y.filename == "" {
		y.filename = parseFileName(y.url)
	}

	fmt.Println("Downloading: ", y.filename)
	length, support, err := checkRangeRequestSupport(y.url)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if !support {
		fmt.Println("Will Implement Single Connection Download Soon")
		return nil, nil
	}

	/*	log.Println("Content-Length: ", length)
		log.Println("Ranged Support: ", support)*/

	filePrefix, _ := generateFileName()

	contentLength, _ := strconv.Atoi(length)

	//chunks := splitFileIntoChunks(414111315, y.ccn)
	chunks := splitFileIntoChunks(contentLength, y.ccn)

	defer cleanUp(y.ccn, filePrefix)

	if len(chunks) == 0 {
		log.Fatalln("an error occurred: couldn't split chunks successfully")
	}

	wg := sync.WaitGroup{}
	stopChan := make(chan struct{})

	wg.Add(y.ccn)

	for index, rng := range chunks {
		go func(i int, r string) {
			defer wg.Done()
			download(filePrefix, strconv.Itoa(i), r, y.url)
		}(index, rng)
	}

	go startSpeedMonitor(filePrefix, contentLength, y.ccn, stopChan)

	wg.Wait()

	log.Println("Finished")
	stopChan <- struct{}{}

	fmt.Println("Consolidating Files Into One File")
	if err = writeFinalFile(y.filename, filePrefix, y.ccn); err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Done")
	return nil, nil
}

func checkRangeRequestSupport(url string) (string, bool, error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return "", false, err
	}

	req.Header.Set("Range", "bytes=0-")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", false, err
	}

	if resp.StatusCode != 200 && resp.StatusCode != 206 {
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

	/*	log.Println(rangeSupport)
		log.Println(contentLength)*/
	if rangeSupport == "bytes" {
		return contentLength, true, nil
	}
	return contentLength, false, nil
}
