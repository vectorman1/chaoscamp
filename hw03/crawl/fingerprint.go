package crawl

import (
	"chaoscamp/hw03/utils"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

type Fingerprint struct {
	URL           string
	ExternalLinks []string
	Technologies  map[string]*Technology
	DateCreated   time.Time
}

type Technology struct {
	Name               string
	DetectedBy         []string
	Categories         []utils.ImportedCategory
	ImportedTechnology utils.ImportedTechnology
}

func NewFingerprint(c *http.Client, link string, s *utils.ScannerData, crawler *Crawler) (*Fingerprint, error) {
	resp, err := crawler.crawl(c, link)
	if err != nil {
		return nil, err
	}

	parsedLink, err := url.Parse(link)
	if err != nil {
		return nil, err
	}

	externalLinks, err := utils.GetExternalLinks(resp.Body)
	if err != nil {
		return nil, err
	}

	f := &Fingerprint{
		URL:           fmt.Sprintf("%s://%s", "https", parsedLink.Host),
		DateCreated:   time.Now(),
		ExternalLinks: externalLinks,
		Technologies:  make(map[string]*Technology),
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := string(bodyBytes)

	for k, v := range s.Technologies {
		f.scanHtml(body, k, &v)
		f.scanHeaders(k, &v, &resp.Header)
		f.scanCookies(k, &v, resp.Cookies())
		f.scanCerts(k, &v, resp)
	}

	return f, nil
}

func (f Fingerprint) uniqueLinks() ([]string, error) {
	if f.ExternalLinks == nil {
		return nil, fmt.Errorf("no external links")
	}

	result := utils.Unique(f.ExternalLinks)

	return result, nil
}

func (f Fingerprint) UnseenUniqueLinks(visited map[string]bool) ([]string, error) {
	uniqueLinks, err := f.uniqueLinks()
	if err != nil {
		return nil, fmt.Errorf("no external links")
	}

	var result []string
	for _, l := range uniqueLinks {
		if v := visited[l]; !v {
			result = append(result, l)
		}
	}

	return result, nil
}

func (f *Fingerprint) technology(t *Technology) {
	if f.Technologies[t.Name] == nil {
		f.Technologies[t.Name] = t
		return
	}

	existingTech := *f.Technologies[t.Name]

	if !utils.Contains(existingTech.DetectedBy, t.DetectedBy[0]) {
		existingTech.DetectedBy = append(existingTech.DetectedBy, t.DetectedBy...)
	}
}

func (f *Fingerprint) scanHeaders(techName string, technology *utils.ImportedTechnology, headers *http.Header) {
	if technology.Headers == nil || headers == nil {
		return
	}

	for hk, hv := range technology.Headers {
		for thk, thv := range *headers {
			if headerKeyMatched, _ := regexp.MatchString(hk, thk); headerKeyMatched {
				if hv == "" {
					tech := Technology{
						Name:               techName,
						DetectedBy:         []string{utils.HEADERS_DETECTED},
						Categories:         utils.GetCategories(*technology),
						ImportedTechnology: *technology,
					}

					f.technology(&tech)
				} else {
					for _, v := range thv {
						if headerValueMatched, _ := regexp.MatchString(hv, v); headerValueMatched {
							tech := Technology{
								Name:               techName,
								DetectedBy:         []string{utils.HEADERS_DETECTED},
								Categories:         utils.GetCategories(*technology),
								ImportedTechnology: *technology,
							}

							f.technology(&tech)
						}
					}
				}
			}
		}
	}

	return
}

func (f *Fingerprint) scanHtml(body string, techName string, technology *utils.ImportedTechnology) {
	if technology.Html == nil {
		return
	}

	for _, htmlRegex := range technology.Html {
		if matched, _ := regexp.MatchString(htmlRegex, body); matched {
			tech := Technology{
				Name:               techName,
				ImportedTechnology: *technology,
				Categories:         utils.GetCategories(*technology),
				DetectedBy:         []string{utils.HTML_DETECTED},
			}

			f.technology(&tech)
		}
	}

	return
}

func (f *Fingerprint) scanCookies(techName string, technology *utils.ImportedTechnology, cookies []*http.Cookie) {
	if technology.Cookies == nil || cookies == nil {
		return
	}

	for tk, tv := range technology.Cookies {
		for _, tc := range cookies {
			if matchedKey, _ := regexp.MatchString(tk, tc.Name); matchedKey {
				if tv == "" {
					tech := Technology{
						Name:               techName,
						DetectedBy:         []string{utils.COOKIES_DETECTED},
						Categories:         utils.GetCategories(*technology),
						ImportedTechnology: *technology,
					}

					f.technology(&tech)
				} else if matchedValue, _ := regexp.MatchString(tv, tc.Value); matchedValue {
					tech := Technology{
						Name:               techName,
						DetectedBy:         []string{utils.COOKIES_DETECTED},
						Categories:         utils.GetCategories(*technology),
						ImportedTechnology: *technology,
					}

					f.technology(&tech)
				}
			}
		}
	}

	return
}

func (f *Fingerprint) scanCerts(techName string, technology *utils.ImportedTechnology, resp *http.Response) {
	if resp.TLS == nil || resp.TLS.PeerCertificates == nil {
		return
	}
	certs := resp.TLS.PeerCertificates

	for _, tcert := range certs {
		for _, name := range tcert.Issuer.Names {
			if name.Value == technology.CertIssuer {
				tech := Technology{
					Name:               techName,
					ImportedTechnology: *technology,
					Categories:         utils.GetCategories(*technology),
					DetectedBy:         []string{utils.CERTS_DETECTED},
				}

				f.technology(&tech)
			}
		}
	}

	return
}

func (f *Fingerprint) saveTechnologies(technologies *chan *Technology, output *chan *Fingerprint) {
	var temp []*Technology

	for t := range *technologies {
		f.technology(t)
		temp = append(temp, t)
	}

	log.Println(temp)

	*output <- f
	return
}
