package yanker

import (
	"bufio"
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

func cleanUp(ccn int, tempFile string) {
	for i := 0; i <= ccn; i++ {
		name := fmt.Sprintf("%d-%s.ynk", i, tempFile)

		os.Remove(name)
	}
}

func writeFinalFile(filename, tempfilename string, ccn int) error {
	file, err := os.Create(filename)

	defer file.Close()

	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	var (
		tempFile *os.File
	)
	for i := 0; i < ccn; i++ {
		tempFile, err = os.Open(fmt.Sprintf("%d-%s.ynk", i, tempfilename))
		buffer := make([]byte, 256)

		if err != nil {
			return err
		}

		reader := bufio.NewReader(tempFile)

		for {
			_, err = reader.Read(buffer)

			if err != nil {
				if err == io.EOF {
					break
				}
			}

			_, err = writer.Write(buffer)

			if err != nil {
				return err
			}

			err = writer.Flush()

			if err != nil {
				return err
			}
		}
		buffer = nil
		func() { _ = tempFile.Close() }()
	}

	return nil
}

func startSpeedMonitor(filename string, contentLength, ccn int, stopChan <-chan struct{}) {
	sizes := make(map[int]int64, ccn)

	for i := 0; i <= ccn; i++ {
		sizes[i] = 0
	}

	for {
		select {
		case <-stopChan:
			return
		case <-time.Tick(time.Second):

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

			fmt.Printf("[%s] - %.2f%% Done - %.2f%s/s           \r", filename, (float64(downloadedSize)/float64(contentLength))*100, speed, sizePrefix)

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
		log.Fatalln(index, " - Not Partial Content")
	}

	defer resp.Body.Close()

	partFile, err := os.OpenFile(fmt.Sprintf("%s-%s.ynk", index, prefix), os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		log.Fatalln(err)
	}

	//	go speedMonitor(partFile.Name(), stop)

	_, err = io.Copy(partFile, resp.Body)
}

func splitFileIntoChunks(size, chunks int) []string {
	/*fileSize, _ := strconv.Atoi(size)
	numChunks, _ := strconv.Atoi(chunks)*/

	if size <= 0 {
		return nil
	}

	chunkSize := size / chunks

	result := make([]string, chunks)

	for i, index := 0, 0; index <= chunks-1; i, index = i+chunkSize, index+1 {
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
