package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"yank/yanker"
)

func main() {
	filename := flag.String("f", "", "-f")
	concurrentConnections := flag.Int("n", 0, "-n")

	flag.Parse()

	url := flag.Args()

	if len(url) == 0 {
		fmt.Printf("Usage: %s [ -f | -n ] <url>\n", os.Args[0])
		return
	}
	var option yanker.Options
	if *filename != "" {
		option.Filename = *filename
	}

	if *concurrentConnections > 0 {
		option.ConcurrentConnections = *concurrentConnections
	}

	y := yanker.NewYankManager(url[0], option)

	if _, err := y.StartDownload(); err != nil {
		log.Fatalln("an error occurred while downloading your file: ", err)
	}
}
