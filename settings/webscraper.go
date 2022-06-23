package settings

import (
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type processSelector func(index int, element *goquery.Selection)

/*
ParseHtmlPage allows parse content of html page
*/
func ParseHtmlPage(url, selector string, processor processSelector) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Printf("Error while trying to GET %s, the error is %s", url, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("error when trying to close url %s: %s", url, err)
			return
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("status code error while trying to get details for url %s: %d %s", url, res.StatusCode, res.Status)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Error while trying to read body from %s, the error is %s", url, err)
		return
	}
	doc.Find(selector).Each(processor)
}
