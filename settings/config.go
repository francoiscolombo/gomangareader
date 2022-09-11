package settings

// Settings is the structure that allowed to store the default configuration and the download history for all the mangas
type Settings struct {
	Config  Config  `json:"config"`
	History History `json:"history"`
}

// Config only store the default configuration, like output path and the global update library flag
type Config struct {
	LibraryPath          string `json:"library_path"`
	AutoUpdate           bool   `json:"auto_update"`
	NbColumns            int    `json:"nb_columns"`
	NbRows               int    `json:"nb_rows"`
	PageWidth            int    `json:"page_width"`
	PageHeight           int    `json:"page_height"`
	ThumbMiniWidth       int    `json:"thumb_mini_width"`
	ThumbMiniHeight      int    `json:"thumb_mini_height"`
	LeftRightButtonWidth int    `json:"left_right_button_width"`
	ChapterLabelWidth    int    `json:"chapter_label_width"`
	ThumbnailWidth       int    `json:"thumbnail_width"`
	ThumbnailHeight      int    `json:"thumbnail_height"`
	ThumbTextHeight      int    `json:"thumb_text_height"`
	NbWorkers            int    `json:"nb_workers"`
}

// History is the manga download history, so it's an array of all the mangas downloaded
type History struct {
	Titles []Manga `json:"titles"`
}

// Manga keep the download history for every manga that we are subscribing
type Manga struct {
	Provider      string    `json:"provider"`
	Title         string    `json:"title"`
	LastChapter   float64   `json:"last_chapter"`
	Chapters      []float64 `json:"chapters"`
	CoverPath     string    `json:"cover_path"`
	Path          string    `json:"path"`
	CoverUrl      string    `json:"cover_url"`
	Name          string    `json:"name"`
	AlternateName string    `json:"alternate_name"`
	YearOfRelease string    `json:"year_of_release"`
	Status        string    `json:"status"`
	Author        string    `json:"author"`
	Artist        string    `json:"artist"`
	Description   string    `json:"description"`
}
