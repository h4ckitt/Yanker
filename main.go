package main

import (
	"yank/yanker"
)

func main() {
	y := yanker.NewYankManager("https://55.download.real-debrid.com/d/SBDLRYUYBWBY6/She-Hulk.Attorney.at.Law.S01E04.720p.WEB.x265-MiNX.mkv", yanker.Options{ConcurrentConnections: 10})
	//y := yanker.NewYankManager("https://www.seedr.cc/download/archive/ef65f774dc8f06171c970d9221bf05a24f597c1211a690974706c079534f6657?token=1fc5b0daf95382c09050ac3149483c31406ddb155886b8916efc28a7fdad91e9&exp=1667052850")
	y.StartDownload()

}
