package main

import (
	"yank/yanker"
)

func main() {
	y := yanker.NewYankManager("https://21.download.real-debrid.com/d/IEPVL5VCVXL5Q/She-Hulk.Attorney.at.Law.S01E09.1080p.HEVC.x265-MeGusta%5Beztv.re%5D.mkv")

	y.StartDownload()

}
