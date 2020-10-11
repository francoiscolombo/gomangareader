package settings

type MangaPanda struct{}

func (provider MangaPanda) FindDetails(libraryPath, title string, lastChapter int) (manga Manga) {
	return Manga{}
}

func (provider MangaPanda) GetPagesUrls(manga Manga) (pageLink []string) {
	var links []string
	return links
}

func (provider MangaPanda) SearchManga(libraryPath, search string) (result []Manga) {
	var titlesFound []Manga
	return titlesFound
}

func (provider MangaPanda) CheckLastChapter(manga Manga) (lastChapter int) {
	return 0
}
