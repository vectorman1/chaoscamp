package main

import (
	"chaoscamp/hw03/crawl"
	"chaoscamp/hw03/utils"
	"flag"
	"fmt"
	"os"
)

func main() {
	recurseFlag := flag.Bool("recurse", true, "Dig into links.")
	maxHeadsFlag := flag.Int("max_heads", 5, "Max concurrent crawls.")
	flag.Parse()

	utils.ReadTechnologies()

	wd, _ := os.Getwd()
	urls, _ := utils.ReadUrls(fmt.Sprintf("%s\\hw03\\urls.txt", wd))

	crawler := crawl.Crawler{
		Sites:          urls,
		WorkList:       make(chan []string),
		UnvisitedLinks: make(chan string),
		VisitedLinks:   make(map[string]bool),
	}

	crawler.Run(recurseFlag, maxHeadsFlag)
}
