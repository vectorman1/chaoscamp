package crawl

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
)

func FetchLinks(r *http.Response) ([]string, error) {
	tree, err := html.Parse(r.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing %s site html %v", r.Request.Host, err)
	}

	var result []string
	visitNode := func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					link, err := r.Request.URL.Parse(a.Val)
					if err != nil {
						continue
					}
					result = append(result, link.String())
				}
			}
		}
	}
	forEachNode(tree, visitNode, nil)

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
