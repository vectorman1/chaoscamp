package crawl

import (
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
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

func (c *Crawler) crawl(client *http.Client, url string) []string {
	log.Println(url)

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "golang_crawler/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	defer resp.Body.Close()

	links, err := FetchLinks(resp)
	if err != nil {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return links
	}

	body := string(bodyBytes)
	emails := make(chan []string)
	phoneNumbers := make(chan []string)
	headerTechnologies := make(chan []string)
	htmlTechnologies := make(chan []string)
	cookiesTechnologies := make(chan []string)
	certTechnologies := make(chan []string)

	go ScanEmails(body, emails)
	go ScanPhoneNumbers(body, phoneNumbers)
	go ScanHtml(body, htmlTechnologies)
	go ScanHeaders(resp.Header, headerTechnologies)
	go ScanCookies(resp.Cookies(), cookiesTechnologies)
	go ScanCerts(resp, certTechnologies)

	go SaveResults(resp, emails, phoneNumbers, headerTechnologies, htmlTechnologies, cookiesTechnologies, certTechnologies)

	return links
}

func SaveResults(
	r *http.Response,
	emails chan []string,
	phoneNumbers chan []string,
	headerTechnologies chan []string,
	htmlTechnologies chan []string,
	cookieTechnologies chan []string,
	certTechnologies chan []string) {
	emailsResult := <-emails
	phoneNumbersResult := <-phoneNumbers
	headerTechnologiesResult := <-headerTechnologies
	htmlTechnologiesResult := <-htmlTechnologies
	cookieTechnologiesResult := <-cookieTechnologies
	certTechnologiesResult := <-certTechnologies

	mergedTechnologies := append(headerTechnologiesResult, append(append(htmlTechnologiesResult, cookieTechnologiesResult...), certTechnologiesResult...)...)
	uniqueTechnologies := unique(mergedTechnologies)

	fmt.Println(r.Request.Host, emailsResult, phoneNumbersResult, uniqueTechnologies)
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

	for i := 0; i < *maxLegs; i++ {
		go func() {
			for link := range c.UnvisitedLinks {
				foundLinks := c.crawl(&client, link)

				if foundLinks == nil {
					fmt.Println("No links found on page.")
				}

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
