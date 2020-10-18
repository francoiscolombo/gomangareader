package settings

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

func contains(array []string, value string) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == value {
			return true
		}
	}
	return false
}

type MangaReader struct{}

func (provider MangaReader) FindDetails(libraryPath, title string, lastChapter int) (manga Manga) {
	// access detail data from mangareader.net only
	// this is working only with this website
	// cover image: present in div.d38
	// description: d46.p
	// details: table.d41
	// - first line: name
	// - second line: alternate name
	// - third line: year of release
	// - forth line: status
	// - fifth line: author
	// - sixth line: artist
	// - seventh line: reading direction
	// - eighth line: genre (a href with class a.d42)
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangareader.net/%s", title), nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error while trying to get details for title %s: %d %s", title, res.StatusCode, res.Status)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	// extract cover url
	var coverUrl string
	doc.Find(".d38").Each(func(i int, div *goquery.Selection) {
		div.Find("img").Each(func(i int, img *goquery.Selection) {
			v, _ := img.Attr("src")
			if strings.HasPrefix(v, "http") {
				coverUrl = v
			} else {
				coverUrl = fmt.Sprintf("https:%s", v)
			}
			return
		})
	})
	// extract description
	var description string
	doc.Find(".d46").Each(func(i int, div *goquery.Selection) {
		div.Find("p").Each(func(i int, p *goquery.Selection) {
			description = p.Text()
			return
		})
	})
	// parse table.d41
	var name string
	var alternateName string
	var yearOfRelease string
	var status string
	var author string
	var artist string
	var readingDirection string
	doc.Find(".d41").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(ir int, row *goquery.Selection) {
			row.Find("td").Each(func(ic int, cell *goquery.Selection) {
				if ic == 1 {
					if ir == 0 {
						name = cell.Text()
					} else if ir == 1 {
						alternateName = cell.Text()
					} else if ir == 2 {
						yearOfRelease = cell.Text()
					} else if ir == 3 {
						status = cell.Text()
					} else if ir == 4 {
						author = cell.Text()
					} else if ir == 5 {
						artist = cell.Text()
					} else if ir == 6 {
						readingDirection = cell.Text()
					}
				}
				return
			})
			return
		})
	})
	metadataPath := filepath.FromSlash(fmt.Sprintf("%s/.metadata", libraryPath))
	coverPath := filepath.FromSlash(fmt.Sprintf("%s/%s-cover.jpg", metadataPath, title))
	// create structure with details to keep
	manga = Manga{
		Provider:         "mangareader.net",
		Title:            title,
		LastChapter:      lastChapter,
		CoverPath:        coverPath,
		CoverUrl:         coverUrl,
		Path:             filepath.FromSlash(fmt.Sprintf("%s/%s", libraryPath, title)),
		Name:             name,
		AlternateName:    alternateName,
		YearOfRelease:    yearOfRelease,
		Status:           status,
		Author:           author,
		Artist:           artist,
		ReadingDirection: readingDirection,
		Description:      description,
	}
	return
}

func (provider MangaReader) GetPagesUrls(manga Manga) (pageLink []string) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangareader.net/%s/%d", manga.Title, manga.LastChapter), nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to get details for title %s: %d %s", manga.Title, res.StatusCode, res.Status)
	}
	body, _ := ioutil.ReadAll(res.Body)
	// we have to analyze the body since the images are located in a script
	content := strings.Split(string(body), "\"")
	for _, val := range content {
		if strings.Contains(val, manga.Title) && strings.HasSuffix(val, ".jpg") {
			link := strings.ReplaceAll(val, "\\/", "/")
			if !strings.HasPrefix(val, "http") {
				link = fmt.Sprintf("https:%s", link)
			}
			if !contains(pageLink, link) {
				pageLink = append(pageLink, link)
			}
		}
	}
	return
}

func (provider MangaReader) SearchManga(libraryPath, search string) (result []Manga) {
	req, _ := http.NewRequest("GET", "https://www.mangareader.net/alphabetical", nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to find title %s: %d %s", search, res.StatusCode, res.Status)
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("a").Each(func(i int, link *goquery.Selection) {
		l, _ := link.Attr("href")
		v := link.Text()
		if strings.HasPrefix(l, "/") {
			l = l[1:]
		}
		if strings.Contains(strings.ToLower(v), strings.ToLower(search)) {
			found := provider.FindDetails(libraryPath, l, 0)
			result = append(result, found)
		}
		return
	})
	return
}

func (provider MangaReader) CheckLastChapter(manga Manga) (lastChapter int) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangareader.net/%s", manga.Title), nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to get details for title %s: %d %s", manga.Title, res.StatusCode, res.Status)
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	// check available chapters
	doc.Find("a").Each(func(i int, link *goquery.Selection) {
		v, _ := link.Attr("href")
		if strings.HasPrefix(v, "/"+manga.Title+"/") {
			s := strings.Split(v, "/")
			lastChapter, _ = strconv.Atoi(s[2])
		}
		return
	})
	return
}
