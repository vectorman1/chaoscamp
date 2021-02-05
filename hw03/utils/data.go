package utils

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"io/ioutil"
	"net/url"
	"regexp"
)

func Unique(slice []string) []string {
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

func Contains(slice []string, t string) bool {
	for _, v := range slice {
		if v == t {
			return true
		}
	}

	return false
}

func GetExternalLinks(body io.Reader) ([]string, error) {
	tree, err := html.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("parsing site html %v", err)
	}

	var result []string
	visitNode := func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					regexMatch, err := regexp.MatchString(URL_REGEX, a.Val)
					if err != nil {
						continue
					}
					if regexMatch {
						link, err := url.Parse(a.Val)
						if err != nil {
							continue
						}
						if link.Hostname() == "" {
							continue
						}
						l := fmt.Sprintf("%s://%s", "https", link.Hostname())
						result = append(result, l)
					}
				}
			}
		}
	}
	forEachNode(tree, visitNode, nil)

	result = Unique(result)
	return result, nil
}

func forEachNode(n *html.Node, pre, post func(n *html.Node)) {
	if pre != nil {
		pre(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEachNode(c, pre, post)
	}
	if post != nil {
		post(n)
	}
}

func SaveToDiskAsJson(val interface{}) {
	bytes, err := json.MarshalIndent(val, "", "\t")
	if err != nil {
		return
	}
	err = ioutil.WriteFile("fingerprints.json", bytes, 0644)
	if err != nil {
		return
	}

	fmt.Println("saved fingerprints results to json")
}
