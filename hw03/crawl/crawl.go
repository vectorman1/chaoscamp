package crawl

import (
	"chaoscamp/hw03/utils"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"
)

type Crawler struct {
	Sites          []string
	WorkList       chan []string
	UnvisitedLinks chan string
	VisitedLinks   map[string]bool
	Fingerprints   []Fingerprint
	Depth          int
}

func (c *Crawler) crawl(client *http.Client, url string) (*http.Response, error) {
	log.Println(url)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "golang_crawler/1.0")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("website responded with %d", resp.StatusCode)
	}

	return resp, nil
}

func (c *Crawler) Run(recurse *bool, maxLegs *int, s *utils.ScannerData) {
	transport := http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		IdleConnTimeout: 5 * time.Second,
	}
	client := http.Client{
		Timeout:   5 * time.Second,
		Transport: &transport,
	}

	go func() { c.WorkList <- c.Sites }()
	var result []Fingerprint
	totalFingerprints := 0
	for i := 0; i < *maxLegs; i++ {
		go func() {
			for link := range c.UnvisitedLinks {
				totalFingerprints++
				fingerprint, err := NewFingerprint(&client, link, s, c)
				if err != nil {
					log.Println("Error generating fingerprint for ", link, err)
					totalFingerprints--
					continue
				}
				links, err := fingerprint.UnseenUniqueLinks(c.VisitedLinks)
				if err != nil {
					log.Println(err)
					totalFingerprints--
					continue
				}
				result = append(result, *fingerprint)

				if *recurse {
					go func() {
						c.WorkList <- links
					}()
				}
			}
		}()
	}

	for list := range c.WorkList {
		if c.Depth == 0 {
			log.Println("waiting for", totalFingerprints-len(result), "fingerprints to finish generating")
			for {
				if len(result) == totalFingerprints {
					break
				}
			}
			fmt.Println("Reached max depth")
			break
		}
		for _, link := range list {
			if !c.VisitedLinks[link] {
				c.VisitedLinks[link] = true
				c.UnvisitedLinks <- link
			}
		}
		c.Depth--
	}

	utils.SaveToDiskAsJson(result)
}
