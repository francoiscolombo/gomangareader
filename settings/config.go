package settings

// Settings is the structure that allowed to store the default configuration and the download history for all the mangas
type Settings struct {
	Config  Config  `json:"config"`
	History History `json:"history"`
}

// Config only store the default configuration, like output path and the global site provider
type Config struct {
	LibraryPath string `json:"library_path"`
}

// History is the manga download history, so it's an array of all the mangas downloaded
type History struct {
	Titles []Manga `json:"titles"`
}

// Manga keep the download history for every mangas that we are subscribing
type Manga struct {
	Provider      string `json:"provider"`
	Title         string `json:"title"`
	LastChapter   int    `json:"last_chapter"`
	CoverPath     string `json:"cover_path"`
	Path          string `json:"path"`
	CoverUrl      string `json:"cover_url"`
	Name          string `json:"name"`
	AlternateName string `json:"alternate_name"`
	YearOfRelease string `json:"year_of_release"`
	Status        string `json:"status"`
	Author        string `json:"author"`
	Artist        string `json:"artist"`
	Description   string `json:"description"`
}
