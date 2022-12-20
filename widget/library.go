package widget

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/francoiscolombo/gomangareader/settings"
)

const (
	versionNumber = "2.1"
	versionName   = "Hōō Genma Ken"
)

// global variables
var config *settings.Settings
var application fyne.App
var mainWindow fyne.Window
var libraryTabs *container.AppTabs
var library *Titles
var details *fyne.Container
var search *Search
var series *Series
var chapters *Chapters
var downloader *Downloader
var reader *Reader

/*
ShowLibrary allow to display the mangas in a GUI.
*/
func ShowLibrary() {

	if settings.IsSettingsExisting() == false {
		settings.WriteDefaultSettings()
	}
	cfg := settings.ReadSettings()
	config = &cfg

	r, _ := fyne.LoadResourceFromPath("./gomangareader.png")
	application = app.NewWithID("gomangareader")
	application.SetIcon(r)

	progress := widget.NewProgressBar()
	mangaTitle := widget.NewLabelWithStyle("...", fyne.TextAlignCenter, fyne.TextStyle{Monospace: true})
	loadLibraryWindow := application.NewWindow(fmt.Sprintf("GoMangaReader v%s (%s)", versionNumber, versionName))
	if config.Config.AutoUpdate == true {
		loadLibraryWindow.SetContent(fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(config.Config.PageWidth, config.Config.ThumbTextHeight)),
			widget.NewLabelWithStyle("Please wait, we are now loading your library, and updating the metadata", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
			widget.NewLabelWithStyle(" at the same time... It could be long, so be patient.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
			progress,
			mangaTitle),
		)
	} else {
		loadLibraryWindow.SetContent(fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(config.Config.PageWidth, config.Config.ThumbTextHeight)),
			widget.NewLabelWithStyle("Please wait, we are now loading your library...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
			progress,
			mangaTitle),
		)
	}
	loadLibraryWindow.CenterOnScreen()
	loadLibraryWindow.Show()

	var searchTab *container.Scroll
	var libraryTab *container.Scroll
	var seriesTab *container.Scroll
	var readerTab *container.Scroll

	go func() {
		mainWindow = application.NewWindow(fmt.Sprintf("GoMangaReader v%s (%s)", versionNumber, versionName))

		library = updateLibraryContent(progress, mangaTitle, config.Config.AutoUpdate)
		libraryTab = container.NewScroll(library)

		search = NewSearch("mangareader.cc")
		searchTab = container.NewScroll(search)

		details = fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
			series,
			chapters,
			downloader,
		)
		seriesTab = container.NewScroll(details)

		reader = nil
		readerTab = container.NewScroll(widget.NewLabel(""))

		configuration := widget.NewButtonWithIcon("Refresh your library...", theme.ViewRefreshIcon(), func() {
			library = updateLibraryContent(progress, mangaTitle, true)
		})

		libraryTabs = container.NewAppTabs(
			container.NewTabItem("Your Library", libraryTab),
			container.NewTabItem("Selected manga", seriesTab),
			container.NewTabItem("Read chapter", readerTab),
			container.NewTabItem("Search new titles", searchTab),
			container.NewTabItem("Preferences", configuration),
		)

		mainWindow.SetContent(libraryTabs)
		mainWindow.Resize(fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns+40, (config.Config.ThumbnailHeight+config.Config.ThumbTextHeight)*config.Config.NbRows+config.Config.ThumbTextHeight))
		mainWindow.SetMaster()
		mainWindow.CenterOnScreen()
		mainWindow.Show()

		loadLibraryWindow.Close()
	}()

	application.Run()
}

func updateLibraryContent(progress *widget.ProgressBar, title *widget.Label, autoUpdate bool) *Titles {
	content := NewTitlesContainer()
	var mangaUpdatedList []settings.Manga
	var provider settings.MangaProvider
	nbTitles := float64(len(config.History.Titles))
	for i, manga := range config.History.Titles {
		value := float64(i) / nbTitles
		title.SetText(manga.Name)
		if autoUpdate {
			provider = settings.MangaReader{}
			newManga := provider.FindDetails(config.Config.LibraryPath, manga.Title, manga.LastChapter)
			provider.BuildChaptersList(&newManga)
			mangaUpdatedList = append(mangaUpdatedList, newManga)
			// download cover picture (if needed)
			err1 := downloadCover(newManga)
			if err1 == nil {
				// and generate thumbnails (if needed)
				err2 := extractFirstPages(config.Config.LibraryPath, newManga)
				if err2 == nil {
					ws := NewTitleButton(manga)
					//groupTitleButtons = append(groupTitleButtons, ws)
					content.Add(ws)
				}
			}
		} else {
			ws := NewTitleButton(manga)
			//groupTitleButtons = append(groupTitleButtons, ws)
			content.Add(ws)
		}
		progress.SetValue(value)
	}
	if autoUpdate {
		// okay we have updated the metadata, now we can save the config
		newSettings := settings.Settings{
			Config: settings.Config{
				LibraryPath:          config.Config.LibraryPath,
				AutoUpdate:           config.Config.AutoUpdate,
				NbColumns:            config.Config.NbColumns,
				NbRows:               config.Config.NbRows,
				PageWidth:            config.Config.PageWidth,
				PageHeight:           config.Config.PageHeight,
				ThumbMiniWidth:       config.Config.ThumbMiniWidth,
				ThumbMiniHeight:      config.Config.ThumbMiniHeight,
				LeftRightButtonWidth: config.Config.LeftRightButtonWidth,
				ChapterLabelWidth:    config.Config.ChapterLabelWidth,
				ThumbnailWidth:       config.Config.ThumbnailWidth,
				ThumbnailHeight:      config.Config.ThumbnailHeight,
				ThumbTextHeight:      config.Config.ThumbTextHeight,
				NbWorkers:            config.Config.NbWorkers,
			},
			History: settings.History{
				Titles: mangaUpdatedList,
			},
		}
		settings.WriteSettings(newSettings)
		//log.Println("> Settings updated.")
	}
	return content
}

func refreshTabsContent(manga *settings.Manga, tabIndex int) {
	series = NewSeries(manga)
	series.Refresh()

	chapters = NewChapters(manga)
	chapters.Refresh()

	downloader = NewDownloader(manga, getLastChapterIndex(*manga))
	downloader.Refresh()

	b := len(manga.Chapters) > 0
	if b {
		b = manga.LastChapter <= manga.Chapters[len(manga.Chapters)-1]
	}
	if b {
		details = fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
			series,
			chapters,
			downloader,
		)
	} else {
		details = fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
			series,
			chapters,
		)
	}
	details.Refresh()

	libraryTab := container.NewScroll(library)
	searchTab := container.NewScroll(search)
	seriesTab := container.NewScroll(details)
	readerTab := container.NewScroll(widget.NewLabel(""))
	if reader != nil {
		readerTab = container.NewScroll(reader)
	}
	configuration := widget.NewLabel("configuration form will be hosted here, if any")

	libraryTabs = container.NewAppTabs(
		container.NewTabItem("Your Library", libraryTab),
		container.NewTabItem("Selected manga", seriesTab),
		container.NewTabItem("Read chapter", readerTab),
		container.NewTabItem("Search new titles", searchTab),
		container.NewTabItem("Preferences", configuration),
	)
	libraryTabs.SelectTabIndex(tabIndex)
	libraryTabs.Refresh()

	mainWindow.SetContent(libraryTabs)
}
