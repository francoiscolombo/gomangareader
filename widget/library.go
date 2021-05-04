package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"log"
)

const (
	versionNumber = "1.0"
	versionName   = "Another Dimension"

	thWidth   = 144
	thHeight  = 192
	chWidth   = 48
	chHeight  = 64
	pgWidth   = 600
	pgHeight  = 585
	btnHeight = 30
)

var globalConfig *settings.Settings

/*
Library allow to display the mangas in a GUI.
*/
func ShowLibrary() {

	libraryViewer := app.NewWithID("gomangareader")
	libraryViewer.SetIcon(theme.FyneLogo())

	windows := libraryViewer.NewWindow("GoMangaReader - Update in progress...")
	windows.SetContent(fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(400, 20)),
		widget.NewLabelWithStyle("Please wait, loading your library...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
		widget.NewProgressBarInfinite()),
	)
	windows.SetMaster()
	windows.CenterOnScreen()
	windows.Show()

	go func() {
		updateLibrary(libraryViewer, windows)
	}()

	libraryViewer.Run()
}

func updateLibrary(app fyne.App, win fyne.Window) {
	win.SetTitle("GoMangaReader - Update in progress...")
	if settings.IsSettingsExisting() == false {
		settings.WriteDefaultSettings()
	}
	cfg := settings.ReadSettings()
	globalConfig = &cfg
	log.Println("- Settings loaded.")
	log.Printf("  > Library path is %s\n", globalConfig.Config.LibraryPath)
	UpdateMetaData(win, *globalConfig)
	cfg = settings.ReadSettings()
	globalConfig = &cfg
	content := fyne.NewContainerWithLayout(layout.NewGridLayout(5))
	for _, manga := range globalConfig.History.Titles {
		ws := widgetSerie(app, win, manga)
		if ws != nil {
			ws.Resize(fyne.NewSize(thWidth, thHeight))
			content.AddObject(ws)
		}
	}
	win.SetContent(fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		widgetUpdateCollections(app, win),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5+40, (thHeight+btnHeight*2)*2)),
			widget.NewScrollContainer(content))))
	win.SetTitle(fmt.Sprintf("GoMangaReader v%s (%s)", versionNumber, versionName))
}

func widgetUpdateCollections(app fyne.App, win fyne.Window) *fyne.Container {
	return fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		widget.NewButtonWithIcon("Search series to add to the collection", theme.SearchIcon(), func() {
			newSerie := widget.NewEntry()
			selectProvider := widget.NewRadio([]string{"mangareader.net", "mangapanda.com"}, func(provider string) {
				log.Println("provider selected:", provider)
			})
			content := widget.NewForm(
				widget.NewFormItem("Which provider to use:", selectProvider),
				widget.NewFormItem("Search series to add:", newSerie),
			)
			dialog.ShowCustomConfirm("What do you want to find?", "Search", "Cancel", content, func(b bool) {
				if !b {
					return
				}
				// and here we have to add it.
				var provider settings.MangaProvider
				if selectProvider.Selected == "mangareader.net" {
					provider = settings.MangaReader{}
				} else if selectProvider.Selected == "mangapanda.com" {
					provider = settings.MangaPanda{}
				}
				searchResults := provider.SearchManga(globalConfig.Config.LibraryPath, newSerie.Text)
				showSearchResults(app, win, newSerie.Text, searchResults)
			}, win)
		}),
		widget.NewButtonWithIcon("Update collection", theme.DocumentIcon(), func() {
			cnf := dialog.NewConfirm("Confirmation", "Are you sure you want to\nupdate your collection?", func(b bool) {
				if !b {
					return
				}
				updateLibrary(app, win)
				canvas.Refresh(win.Canvas().Content())
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "GoMangaReader",
					Content: "Your library is now updated.",
				})
			}, win)
			cnf.SetDismissText("Nah")
			cnf.SetConfirmText("Oh Yes!")
			cnf.Show()
		}),
	)
}

func widgetSerie(app fyne.App, win fyne.Window, manga settings.Manga) *fyne.Container {
	title := manga.Name
	if len(title) > 13 {
		title = title[0:13] + "..."
	}
	colorTitle := color.NRGBA{0x80, 0xff, 0, 0xff}
	if checkNewChapters(manga) {
		colorTitle = color.NRGBA{0xff, 0x80, 0, 0xff}
	}
	return fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, thHeight)), canvas.NewImageFromFile(manga.CoverPath)),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, 20)), canvas.NewText(fmt.Sprintf("%s (%d)", title, manga.LastChapter-1), colorTitle)),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, btnHeight)), widget.NewButton("Show", func() {
			showSerieDetail(app, manga)
		})))
}
