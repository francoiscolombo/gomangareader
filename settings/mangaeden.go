package settings

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/francoiscolombo/gomangareader/webscraper"
)

type MangaDen struct{}

var manga = Manga{}
var pageLink []string
var searchResults []Manga
var lastChapter = 0

func extractMangaProperties(index int, element *goquery.Selection) {

}

func extractPageLinks(index int, element *goquery.Selection) {

}

func (provider MangaDen) FindDetails(libraryPath, title string, lastChapter int) (manga Manga) {
	webscraper.ParseHtmlPage(fmt.Sprintf("%s", title), "#rightBox", extractMangaProperties)
}

func (provider MangaDen) GetPagesUrls(manga Manga) (pageLink []string) {

}

func (provider MangaDen) SearchManga(libraryPath, search string) (result []Manga) {

}

func (provider MangaDen) CheckLastChapter(manga Manga) (lastChapter int) {

}
