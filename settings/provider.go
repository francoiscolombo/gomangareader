package settings

// MangaProvider is the interface for allowing to use more than one provider
type MangaProvider interface {
	FindDetails(libraryPath, title string, lastChapter float64) (manga Manga)
	GetPagesUrls(manga Manga) (pageLink []string)
	SearchManga(libraryPath, search string) (result []Manga)
	CheckLastChapter(manga Manga) (lastChapter float64)
	BuildChaptersList(manga *Manga) Manga
}
