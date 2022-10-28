package yanker

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"time"
)

const (
	_ = 1 << (iota * 10)
	KB
	MB
	GB
	TB
)

const LETTERS = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

func getSize(fileName string) (int64, error) {
	info, err := os.Stat(fileName)

	if err != nil {
		return -1, err
	}

	return info.Size(), nil

}

func speedMonitor(filename string, stopChan chan struct{}) {
	prevSize := int64(0)

	for {
		select {
		case <-stopChan:
			return
		default:
			size, _ := getSize(filename)

			diff := size - prevSize
			prevSize = size
			var speed float64
			sizePrefix := ""
			switch {
			case diff >= TB:
				speed = float64(diff) / TB
				sizePrefix = "TB"
			case diff >= GB:
				speed = float64(diff) / GB
				sizePrefix = "GB"
			case diff >= MB:
				speed = float64(diff) / MB
				sizePrefix = "MB"
			case diff >= KB:
				speed = float64(diff) / KB
				sizePrefix = "KB"
			default:
				speed = float64(diff)
				sizePrefix = "B"
			}

			fmt.Printf("%s - %.2f%s/s\r", filename, speed, sizePrefix)

		}
	}
}

func startSpeedMonitor(filename string, ccn int, stopChan chan struct{}) {
	sizes := make(map[int]int64, ccn)

	for i := 0; i <= ccn; i++ {
		sizes[i] = 0
	}

	for {
		select {
		case <-stopChan:
			return
		default:
			for range time.Tick(time.Second) {
				var (
					previousDownloadedSize int64
					downloadedSize         int64
				)
				for i := 0; i <= ccn; i++ {
					tempFileName := fmt.Sprintf("%d-%s.ynk", i, filename)
					size, _ := getSize(tempFileName)
					downloadedSize += size
					previousDownloadedSize += sizes[i]
					sizes[i] = size
				}

				diff := downloadedSize - previousDownloadedSize
				var speed float64
				sizePrefix := ""
				switch {
				case diff >= TB:
					speed = float64(diff) / TB
					sizePrefix = "TB"
				case diff >= GB:
					speed = float64(diff) / GB
					sizePrefix = "GB"
				case diff >= MB:
					speed = float64(diff) / MB
					sizePrefix = "MB"
				case diff >= KB:
					speed = float64(diff) / KB
					sizePrefix = "KB"
				default:
					speed = float64(diff)
					sizePrefix = "B"
				}
				fmt.Printf("%s - %.2f%s/s\r", filename, speed, sizePrefix)
			}
		}
	}

}

func download(prefix, index, bRange, url string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Fatalln(err)
	}
	byteRange := fmt.Sprintf("bytes=%s", bRange)
	req.Header.Set("Range", byteRange)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	if resp.StatusCode != http.StatusPartialContent {
		log.Fatalln("Not Partial Content")
	}

	defer resp.Body.Close()

	partFile, err := os.OpenFile(fmt.Sprintf("%s-%s.ynk", index, prefix), os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatalln(err)
	}

	stop := make(chan struct{})

	//	go speedMonitor(partFile.Name(), stop)

	_, err = io.Copy(partFile, resp.Body)
	stop <- struct{}{}
}

func splitFileIntoChunks(size, chunks int) []string {
	/*fileSize, _ := strconv.Atoi(size)
	numChunks, _ := strconv.Atoi(chunks)*/

	log.Println(chunks)

	chunkSize := size / chunks

	log.Println(chunkSize)

	result := make([]string, chunks)

	for i, index := 0, 0; index <= chunks-1; i, index = i+chunkSize, index+1 {
		log.Println(index)
		result[index] = fmt.Sprintf("%d-%d", i, (i+chunkSize)-1)
	}

	last := chunks - 1
	result[last] = fmt.Sprintf("%d-%d", last*chunkSize, ((last*chunkSize)+chunkSize)-1)

	return result
}

func generateFileName() (string, error) {
	ret := make([]byte, 8)

	for i := 0; i < 8; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(LETTERS))))

		if err != nil {
			log.Println(err)
			return "", err
		}
		ret[i] = LETTERS[num.Int64()]
	}

	return base64.URLEncoding.EncodeToString(ret), nil
}
