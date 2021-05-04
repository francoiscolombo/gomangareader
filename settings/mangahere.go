package settings

type MangaHere struct{}

func (provider MangaHere) FindDetails(libraryPath, title string, lastChapter int) (manga Manga) {

}

func (provider MangaHere) GetPagesUrls(manga Manga) (pageLink []string) {

}

func (provider MangaHere) SearchManga(libraryPath, search string) (result []Manga) {

}

func (provider MangaHere) CheckLastChapter(manga Manga) (lastChapter int) {

}
