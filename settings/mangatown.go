package settings

type MangaTown struct{}

func (provider MangaTown) FindDetails(libraryPath, title string, lastChapter int) (manga Manga) {

}

func (provider MangaTown) GetPagesUrls(manga Manga) (pageLink []string) {

}

func (provider MangaTown) SearchManga(libraryPath, search string) (result []Manga) {

}

func (provider MangaTown) CheckLastChapter(manga Manga) (lastChapter int) {

}
