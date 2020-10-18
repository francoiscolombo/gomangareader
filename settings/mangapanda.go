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

type MangaPanda struct{}

func (provider MangaPanda) FindDetails(libraryPath, title string, lastChapter int) (manga Manga) {
	// access detail data from mangapanda.com only
	// this is working only with this website
	// cover image: present in div.mangaimg
	// properties in div.mangaproperties
	// description: div.readmangasum.p
	// details: in a table inside div.mangaproperties, 1st column: property title, second column: value
	// - name: 1st line
	// - alternate name: 2nd line
	// - year of release: 3rd line
	// - status: 4th line
	// - author: 5th line
	// - artist: 6th line
	// - reading direction: 7th line
	// - genre: 8th line, inside span.genretags
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangapanda.com/%s", title), nil)
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
	// we are going to extract all of these
	var coverUrl string
	var description string
	var name string
	var alternateName string
	var yearOfRelease string
	var status string
	var author string
	var artist string
	var readingDirection string
	// coverurl
	doc.Find("#mangaimg").Each(func(i int, div *goquery.Selection) {
		div.Find("img").Each(func(i int, img *goquery.Selection) {
			v, _ := img.Attr("src")
			if strings.HasPrefix(v, "http") {
				coverUrl = v
			} else {
				coverUrl = fmt.Sprintf("https:%s", v)
			}
		})
	})
	// parse table inside div.mangaproperties
	doc.Find("#mangaproperties").Each(func(i int, div *goquery.Selection) {
		div.Find("table").Each(func(i int, table *goquery.Selection) {
			table.Find("tr").Each(func(ir int, row *goquery.Selection) {
				row.Find("td").Each(func(ic int, cell *goquery.Selection) {
					if ic == 1 {
						if ir == 0 {
							name = strings.TrimSpace(cell.Text())
						} else if ir == 1 {
							alternateName = strings.TrimSpace(cell.Text())
						} else if ir == 2 {
							yearOfRelease = strings.TrimSpace(cell.Text())
						} else if ir == 3 {
							status = strings.TrimSpace(cell.Text())
						} else if ir == 4 {
							author = strings.TrimSpace(cell.Text())
						} else if ir == 5 {
							artist = strings.TrimSpace(cell.Text())
						} else if ir == 6 {
							readingDirection = strings.TrimSpace(cell.Text())
						}
					}
					return
				})
				return
			})
		})
	})
	// extract description
	doc.Find("#readmangasum").Each(func(i int, div *goquery.Selection) {
		div.Find("p").Each(func(i int, p *goquery.Selection) {
			description = p.Text()
			return
		})
	})
	metadataPath := filepath.FromSlash(fmt.Sprintf("%s/.metadata", libraryPath))
	coverPath := filepath.FromSlash(fmt.Sprintf("%s/%s-cover.jpg", metadataPath, title))
	// create structure with details to keep
	manga = Manga{
		Provider:         "mangapanda.com",
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

func (provider MangaPanda) GetPagesUrls(manga Manga) (pageLink []string) {
	// for this site, we have one page showing at a time, with a link to the next page until the end of the chapter,
	// which link to the next chapter instead.
	// the image is located inside a div.imgholder, which contains the a.href with the link to the next page or next
	// chapter, and inside the a with have a img (with id img but this is not relevant)
	// so we have to iterate, until we reach the next chapter.
	// we have to analyze the body since the images are located in a script
	chapterLink := fmt.Sprintf("https://www.mangapanda.com/%s/%d", manga.Title, manga.LastChapter)
	nextLink := chapterLink
	for {
		req, _ := http.NewRequest("GET", nextLink, nil)
		req.Header.Add("cache-control", "no-cache")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			// handle error
			log.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Fatalf("status code error while trying to load pages for title %s: %d %s", manga.Title, res.StatusCode, res.Status)
		}
		body, _ := ioutil.ReadAll(res.Body)
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
		if err != nil {
			log.Fatal(err)
		}
		haveNextPage := true
		doc.Find("#imgholder").Each(func(i int, div *goquery.Selection) {
			div.Find("a").Each(func(i int, link *goquery.Selection) {
				nextLink, _ = link.Attr("href")
				nextLink = fmt.Sprintf("https://www.mangapanda.com%s",nextLink)
				haveNextPage = strings.HasPrefix(nextLink, chapterLink)
				log.Println("next link:",nextLink," have next page? ",haveNextPage)
				link.Find("img").Each(func(i int, img *goquery.Selection) {
					imgLink, _ := img.Attr("src")
					log.Println("found image link:",imgLink)
					if !contains(pageLink, imgLink) {
						pageLink = append(pageLink, imgLink)
					}
				})
			})
		})
		if !haveNextPage {
			break
		}
	}
	log.Println("links found:",pageLink)
	return
}

func (provider MangaPanda) SearchManga(libraryPath, search string) (result []Manga) {
	// we are going to search on the alphabetical list of the available mangas.
	// no search in description on every single manga since it is too much time consuming
	// we are just search in the text value of all the a.href starting with a /
	req, _ := http.NewRequest("GET", "https://www.mangapanda.com/alphabetical", nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to find title %s: %d %s", search, res.StatusCode, res.Status)
		return
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
	})
	return
}

func (provider MangaPanda) CheckLastChapter(manga Manga) (lastChapter int) {
	// the last chapter is available from the detail page, and its located in a div.latestchapters, and it is the first
	// a.href. the link is on this format: /<title>/<chapter number>, so we need split on "/" and extract the item #2
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangapanda.com/%s", manga.Title), nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error while trying to get details for title %s: %d %s", manga.Title, res.StatusCode, res.Status)
		return
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("#latestchapters").Each(func(i int, div *goquery.Selection) {
		div.Find("a").Each(func(i int, link *goquery.Selection) {
			v, _ := link.Attr("href")
			if i == 0 {
				vv := strings.Split(v,"/")
				fmt.Println("split>",vv)
				lastChapter, err = strconv.Atoi(vv[2])
				if err != nil {
					log.Fatal(err)
				}
				return
			}
		})
	})
	return
}
