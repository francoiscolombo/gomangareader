package settings

// MangaProvider is the interface for allowing to use more than one provider
type MangaProvider interface {
	FindDetails(libraryPath, title string, lastChapter int) (manga Manga)
	GetPagesUrls(manga Manga) (pageLink []string)
	SearchManga(libraryPath, search string) (result []Manga)
	CheckLastChapter(manga Manga) (lastChapter int)
}
