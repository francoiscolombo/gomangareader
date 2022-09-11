package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/settings"
)

const (
	versionNumber = "2.1"
	versionName   = "Hōō Genma Ken"
)

var globalConfig *settings.Settings
var mainWindow fyne.Window
var titles *Titles

/*
ShowLibrary allow to display the mangas in a GUI.
*/
func ShowLibrary() {

	if settings.IsSettingsExisting() == false {
		settings.WriteDefaultSettings()
	}
	cfg := settings.ReadSettings()
	globalConfig = &cfg

	r, _ := fyne.LoadResourceFromPath("./gomangareader.png")
	libraryViewer := app.NewWithID("gomangareader")
	libraryViewer.SetIcon(r)

	progress := widget.NewProgressBar()
	mangaTitle := widget.NewLabelWithStyle("...", fyne.TextAlignCenter, fyne.TextStyle{Monospace: true})
	winload := libraryViewer.NewWindow(fmt.Sprintf("GoMangaReader v%s (%s)", versionNumber, versionName))
	if globalConfig.Config.AutoUpdate == true {
		winload.SetContent(fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(globalConfig.Config.PageWidth, globalConfig.Config.ThumbTextHeight)),
			widget.NewLabelWithStyle("Please wait, we are now loading your library, and updating the metadata", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
			widget.NewLabelWithStyle(" at the same time... It could be long, so be patient.", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
			progress,
			mangaTitle),
		)
	} else {
		winload.SetContent(fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(globalConfig.Config.PageWidth, globalConfig.Config.ThumbTextHeight)),
			widget.NewLabelWithStyle("Please wait, we are now loading your library...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
			progress,
			mangaTitle),
		)
	}
	winload.CenterOnScreen()
	winload.Show()

	var searchTitles *container.Scroll
	var library *container.Scroll

	go func() {
		mainWindow = libraryViewer.NewWindow(fmt.Sprintf("GoMangaReader v%s (%s)", versionNumber, versionName))

		titles = updateLibraryContent(libraryViewer, progress, mangaTitle, globalConfig.Config.AutoUpdate)
		library = container.NewScroll(titles)

		search := NewSearch(libraryViewer, "mangareader.cc")
		searchTitles = container.NewScroll(search)

		configuration := widget.NewLabel("configuration form will be hosted here, if any")

		mainWindow.SetContent(container.NewAppTabs(
			container.NewTabItem("Your Library", library),
			container.NewTabItem("Search new titles", searchTitles),
			container.NewTabItem("Preferences", configuration),
		))
		mainWindow.Resize(fyne.NewSize(globalConfig.Config.ThumbnailWidth*globalConfig.Config.NbColumns+40, (globalConfig.Config.ThumbnailHeight+globalConfig.Config.ThumbTextHeight)*globalConfig.Config.NbRows+globalConfig.Config.ThumbTextHeight))
		mainWindow.SetMaster()
		mainWindow.CenterOnScreen()
		mainWindow.Show()

		winload.Close()
	}()

	libraryViewer.Run()
}

func updateLibraryContent(app fyne.App, progress *widget.ProgressBar, title *widget.Label, autoUpdate bool) *Titles {
	content := NewTitlesContainer(app)
	var mangaUpdatedList []settings.Manga
	var provider settings.MangaProvider
	nbTitles := float64(len(globalConfig.History.Titles))
	for i, manga := range globalConfig.History.Titles {
		value := float64(i) / nbTitles
		title.SetText(manga.Name)
		if autoUpdate {
			provider = settings.MangaReader{}
			newManga := provider.FindDetails(globalConfig.Config.LibraryPath, manga.Title, manga.LastChapter)
			provider.BuildChaptersList(&newManga)
			mangaUpdatedList = append(mangaUpdatedList, newManga)
			// download cover picture (if needed)
			err1 := downloadCover(newManga)
			if err1 == nil {
				// and generate thumbnails (if needed)
				err2 := extractFirstPages(globalConfig.Config.LibraryPath, newManga)
				if err2 == nil {
					ws := NewTitleButton(app, content, manga)
					//groupTitleButtons = append(groupTitleButtons, ws)
					content.Add(ws)
				}
			}
		} else {
			ws := NewTitleButton(app, content, manga)
			//groupTitleButtons = append(groupTitleButtons, ws)
			content.Add(ws)
		}
		progress.SetValue(value)
	}
	if autoUpdate {
		// okay we have updated the metadata, now we can save the config
		newSettings := settings.Settings{
			Config: settings.Config{
				LibraryPath:          globalConfig.Config.LibraryPath,
				AutoUpdate:           globalConfig.Config.AutoUpdate,
				NbColumns:            globalConfig.Config.NbColumns,
				NbRows:               globalConfig.Config.NbRows,
				PageWidth:            globalConfig.Config.PageWidth,
				PageHeight:           globalConfig.Config.PageHeight,
				ThumbMiniWidth:       globalConfig.Config.ThumbMiniWidth,
				ThumbMiniHeight:      globalConfig.Config.ThumbMiniHeight,
				LeftRightButtonWidth: globalConfig.Config.LeftRightButtonWidth,
				ChapterLabelWidth:    globalConfig.Config.ChapterLabelWidth,
				ThumbnailWidth:       globalConfig.Config.ThumbnailWidth,
				ThumbnailHeight:      globalConfig.Config.ThumbnailHeight,
				ThumbTextHeight:      globalConfig.Config.ThumbTextHeight,
				NbWorkers:            globalConfig.Config.NbWorkers,
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
