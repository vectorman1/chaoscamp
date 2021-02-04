package crawl

import (
	"chaoscamp/hw03/utils"
	"log"
	"net/http"
	"regexp"
)

var (
	scanPhoneNumberRegExpr = regexp.MustCompile("([+]*[(]?[0-9]{1,4}[)]?[-\\s./0-9]*[^\\D])")
	scanEmailRegExpr       = regexp.MustCompile("(?i)([A-Z0-9._%+-]+@[A-Z0-9.-]+\\\\.[A-Z]{2,24})")
)

func ScanHeaders(headers http.Header, result chan []string) {
	r := make([]string, 0)

	scannerData := utils.GetScannerData()

	for k, t := range scannerData.Technologies {
		for hk, hv := range t.Headers {
			for thk, thv := range headers {
				if headerKeyMatched, _ := regexp.MatchString(hk, thk); headerKeyMatched {
					if hv != "" {
						for _, v := range thv {
							if matched, _ := regexp.MatchString(hv, v); matched {
								r = append(r, k)
								if len(t.Implies) > 0 {
									r = append(r, t.Implies...)
								}
							}
						}
					} else {
						r = append(r, k)
						if len(t.Implies) > 0 {
							r = append(r, t.Implies...)
						}
					}
				}
			}
		}
	}

	result <- r

	defer close(result)
}
func ScanHtml(body string, result chan []string) {
	if body == "" {
		result <- nil
		log.Println("Body was empty")
		return
	}

	r := make([]string, 0)

	scannerData := utils.GetScannerData()

	for k, t := range scannerData.Technologies {
		for _, htmlRegex := range t.Html {
			matched, _ := regexp.MatchString(htmlRegex, body)
			if matched {
				r = append(r, k)
				if len(t.Implies) > 0 {
					r = append(r, t.Implies...)
				}
			}
		}
	}

	result <- r

	defer close(result)
}
func ScanPhoneNumbers(body string, result chan []string) {
	values := scanPhoneNumberRegExpr.FindStringSubmatch(body)

	result <- values

	defer close(result)
}
func ScanEmails(body string, result chan []string) {
	values := scanEmailRegExpr.FindStringSubmatch(body)

	result <- values

	defer close(result)
}
func ScanCookies(cookies []*http.Cookie, result chan []string) {
	var r []string

	scannerData := utils.GetScannerData()

	for k, t := range scannerData.Technologies {
		for c, v := range t.Cookies {
			for _, tc := range cookies {
				if c == tc.Name {
					if v != "" {
						matches, _ := regexp.MatchString(v, tc.Value)
						if matches {
							r = append(r, k)
							if len(t.Implies) > 0 {
								r = append(r, t.Implies...)
							}
						}
					} else {
						r = append(r, k)
						if len(t.Implies) > 0 {
							r = append(r, t.Implies...)
						}
					}
				}
			}
		}
	}

	result <- r

	defer close(result)
}
