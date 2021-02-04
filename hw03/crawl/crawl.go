package crawl

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
)

type Crawler struct {
	mux            sync.Mutex
	Sites          []string
	Out            io.Writer
	WorkList       chan []string
	UnvisitedLinks chan string
	VisitedLinks   map[string]bool
}

type Fetcher interface {
	Fetch(url string) (body string, urls []string, err error)
}

func (c *Crawler) crawl(url string) []string {
	log.Println(url)

	r, err := http.Get(url)
	if err != nil {
		return nil
	}
	if r.StatusCode != http.StatusOK {
		_ = r.Body.Close()
		return nil
	}

	links, err := FetchLinks(r)
	if err != nil {
		return nil
	}
	if r.StatusCode != http.StatusOK {
		_ = r.Body.Close()
		return nil
	}

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return links
	}

	body := string(bodyBytes)
	emails := make(chan []string)
	phoneNumbers := make(chan []string)
	headerTechnologies := make(chan []string)
	htmlTechnologies := make(chan []string)
	cookiesTechnologies := make(chan []string)

	go ScanEmails(body, emails)
	go ScanPhoneNumbers(body, phoneNumbers)
	go ScanHtml(body, htmlTechnologies)
	go ScanHeaders(r.Header, headerTechnologies)
	go ScanCookies(r.Cookies(), cookiesTechnologies)

	go SaveResults(r, emails, phoneNumbers, headerTechnologies, htmlTechnologies, cookiesTechnologies)

	return links
}

func SaveResults(r *http.Response, emails chan []string, phoneNumbers chan []string, headerTechnologies chan []string, htmlTechnologies chan []string, cookieTechnologies chan []string) {
	emailsResult := <-emails
	phoneNumbersResult := <-phoneNumbers
	headerTechnologiesResult := <-headerTechnologies
	htmlTechnologiesResult := <-htmlTechnologies
	cookieTechnologiesResult := <-cookieTechnologies
	mergedTechnologies := append(headerTechnologiesResult, append(htmlTechnologiesResult, cookieTechnologiesResult...)...)
	uniqueTechnologies := unique(mergedTechnologies)

	fmt.Println(emailsResult, phoneNumbersResult, uniqueTechnologies)
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func (c *Crawler) Run(recurse *bool, maxLegs *int) {
	go func() { c.WorkList <- c.Sites }()

	for i := 0; i < *maxLegs; i++ {
		go func() {
			for link := range c.UnvisitedLinks {
				foundLinks := c.crawl(link)

				if *recurse {
					go func() {
						c.WorkList <- foundLinks
					}()
				}
			}
		}()
	}

	for list := range c.WorkList {
		for _, link := range list {
			pLink, err := url.Parse(link)
			if err != nil {
				fmt.Errorf("Error parsing url found %s", link)
				continue
			}
			if !c.VisitedLinks[pLink.Hostname()] {
				c.VisitedLinks[pLink.Hostname()] = true
				c.UnvisitedLinks <- link
			}
		}
	}
}
