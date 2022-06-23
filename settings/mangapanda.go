package settings

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

const MangaPandaSiteUrl = "http://www.mangapanda.in"

type MangaPanda struct{}

func (provider MangaPanda) FindDetails(libraryPath, title string, lastChapter int) (manga Manga) {
	// access detail data from mangapanda.in only
	// this is working only with this website
	// cover image: present in div.cover-detail
	// name: h1.title-manga
	// remaining properties are inside a p.description-update tag
	// for the description, we have to find div.manga-content and get the content of the embedded p tag
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/manga/%s", MangaPandaSiteUrl, title), nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Printf("Can't GET %s/manga/%s, error is %s", MangaPandaSiteUrl, title, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Something went wrong while trying to close the http client, error is %s", err)
			return
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("status code error while trying to get details for title %s: %d %s", title, res.StatusCode, res.Status)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Can't read body from %s/manga/%s, error is %s", MangaPandaSiteUrl, title, err)
		return
	}
	// we are going to extract all of these
	var coverUrl string
	var description string
	var name string
	var properties string
	var alternate string
	var author string
	var artist string
	var release string
	var status string
	// coverurl
	doc.Find(".cover-detail").Each(func(i int, div *goquery.Selection) {
		div.Find("img").Each(func(i int, img *goquery.Selection) {
			v, _ := img.Attr("src")
			if strings.HasPrefix(v, "http") {
				coverUrl = v
				return
			} else {
				coverUrl = fmt.Sprintf("https:%s", v)
				return
			}
		})
	})
	// search name
	doc.Find(".title-manga").Each(func(i int, h1 *goquery.Selection) {
		name = h1.Text()
		return
	})
	// extract properties
	doc.Find(".description-update").Each(func(i int, p *goquery.Selection) {
		properties = p.Text()
		return
	})
	for _, line := range strings.Split(properties, "\n") {
		fields := strings.Split(line, ":")
		if len(fields) == 2 {
			k := strings.ToLower(strings.TrimSpace(fields[0]))
			v := strings.TrimSpace(fmt.Sprintf("%s", fields[1]))
			if strings.HasPrefix(k, "alternative") {
				alternate = v
			} else if strings.HasPrefix(k, "author") {
				author = v
			} else if strings.HasPrefix(k, "artist") {
				artist = v
			} else if strings.HasPrefix(k, "status") {
				status = v
			} else if strings.HasPrefix(k, "release") {
				release = v
			}
		}
	}
	// extract description
	doc.Find(".manga-content").Each(func(i int, div *goquery.Selection) {
		div.Find("p").Each(func(i int, p *goquery.Selection) {
			description = p.Text()
			return
		})
	})
	metadataPath := filepath.FromSlash(fmt.Sprintf("%s/.metadata", libraryPath))
	coverPath := filepath.FromSlash(fmt.Sprintf("%s/%s-cover.jpg", metadataPath, title))
	// create structure with details to keep
	manga = Manga{
		Provider:      "mangapanda.in",
		Title:         title,
		LastChapter:   lastChapter,
		CoverPath:     coverPath,
		CoverUrl:      coverUrl,
		Path:          filepath.FromSlash(fmt.Sprintf("%s/%s", libraryPath, title)),
		Name:          name,
		AlternateName: alternate,
		YearOfRelease: release,
		Status:        status,
		Author:        author,
		Artist:        artist,
		Description:   strings.TrimSpace(description),
	}
	return
}

func (provider MangaPanda) GetPagesUrls(manga Manga) (pageLink []string) {
	// for this site, we have all the pages for a chapter listed inside the body content
	// they are located under a tag <p id=arraydata style=display:none>...</p> as a list
	// separated by commas. so we just need to query this id, get the content, split the
	// string by commas then we get our list of images.
	// format of the url >
	chapterLink := fmt.Sprintf("%s/%s-chapter-%d#1", MangaPandaSiteUrl, manga.Title, manga.LastChapter)
	req, _ := http.NewRequest("GET", chapterLink, nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Printf("Can't GET %s/%s-chapter-%d#1, error is %s", MangaPandaSiteUrl, manga.Title, manga.LastChapter, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Something went wrong while trying to close the http client, error is %s", err)
			return
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("status code error while trying to load pages for title %s: %d %s", manga.Title, res.StatusCode, res.Status)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Can't read body from %s/%s-chapter-%d#1, error is %s", MangaPandaSiteUrl, manga.Title, manga.LastChapter, err)
		return
	}
	doc.Find("#arraydata").Each(func(i int, p *goquery.Selection) {
		content := p.Text()
		pageLink = strings.Split(content, ",")
	})
	log.Println("links found:", pageLink)
	return
}

func (provider MangaPanda) SearchManga(libraryPath, search string) (result []Manga) {
	// here to search mangas, we need to use the following query:
	// http://mangapanda.in/search?q=<text>
	prm := url.Values{}
	prm.Add("q", search)
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/search?%s", MangaPandaSiteUrl, prm.Encode()), nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Printf("Can't GET %s/search?%s, error is %s", MangaPandaSiteUrl, prm.Encode(), err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Something went wrong while trying to close the http client, error is %s", err)
			return
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("status code error while trying to find title %s: %d %s", search, res.StatusCode, res.Status)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Can't read body from %s/search?%s, error is %s", MangaPandaSiteUrl, prm.Encode(), err)
		return
	}
	doc.Find(".media-body").Each(func(i int, div *goquery.Selection) {
		//log.Printf("SEARCH:> i:%d, selection is %s", i, div.Text())
		div.Find("a").Each(func(i int, link *goquery.Selection) {
			//log.Printf("SEARCH:> i:%d, embeded link is %s", i, link.Text())
			l, _ := link.Attr("href")
			log.Printf("SEARCH:> i:%d, embeded link ref is %s", i, l)
			v := link.Text()
			if strings.HasPrefix(l, "/") {
				l = l[1:]
			}
			if strings.Contains(strings.ToLower(v), strings.ToLower(search)) {
				ll := strings.Split(l, "/")
				if len(ll) > 0 {
					l = ll[len(ll)-1]
				}
				log.Printf("SEARCH:> this is a match, get details for title %s...", l)
				found := provider.FindDetails(libraryPath, l, 0)
				//log.Printf("SEARCH:> found %s", found)
				result = append(result, found)
			}
		})
	})
	return
}

func (provider MangaPanda) CheckLastChapter(manga Manga) (lastChapter int) {
	// the last chapter is available from the detail page, and its located in a div.latestchapters, and it is the first
	// a.href.
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/manga/%s", MangaPandaSiteUrl, manga.Title), nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Printf("Can't request %s, error is %s", manga.Provider, err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Something went wrong while trying to close the http client, error is %s", err)
			return
		}
	}(res.Body)
	if res.StatusCode != 200 {
		log.Printf("status code error while trying to get details for title %s: %d %s", manga.Title, res.StatusCode, res.Status)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Printf("Error while trying to retrieve manga %s, error is %s", manga.Title, err)
		return
	}
	lastChapter = -1
	doc.Find(".chapter-list").Each(func(i int, div *goquery.Selection) {
		div.Find("a").Each(func(i int, link *goquery.Selection) {
			if lastChapter == -1 {
				v, _ := link.Attr("href")
				vv := strings.Split(v, "/")
				vvv := strings.Split(vv[len(vv)-1], "-")
				fmt.Println("split>", vvv)
				lastChapter, err = strconv.Atoi(vvv[len(vvv)-1])
				if err != nil {
					log.Printf("Error while trying to get last chapter > %s", err)
					lastChapter = -1
				} else {
					return
				}
			}
		})
		if lastChapter > 0 {
			return
		}
	})
	return
}
