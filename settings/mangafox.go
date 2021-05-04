package settings

type MangaFox struct{}

func (provider MangaFox) FindDetails(libraryPath, title string, lastChapter int) (manga Manga) {

}

func (provider MangaFox) GetPagesUrls(manga Manga) (pageLink []string) {

}

func (provider MangaFox) SearchManga(libraryPath, search string) (result []Manga) {

}

func (provider MangaFox) CheckLastChapter(manga Manga) (lastChapter int) {

}
