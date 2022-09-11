package widget

import (
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/archive"
	"github.com/francoiscolombo/gomangareader/settings"
	"os"
	"path/filepath"
)

func isMangaChapterExists(manga settings.Manga, chapter float64) bool {
	cbzPath := filepath.FromSlash(fmt.Sprintf("%s/%s-%03.1f.cbz", manga.Path, manga.Title, chapter))
	info, err := os.Stat(cbzPath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func readChapter(app fyne.App, manga settings.Manga, chapter float64) {
	if isMangaChapterExists(manga, chapter) {
		w := app.NewWindow(fmt.Sprintf("Read %s - chapter %.1f", manga.Name, chapter))
		w.SetContent(container.NewScroll(widgetReader(app, manga, chapter)))
		w.Resize(fyne.NewSize(globalConfig.Config.PageWidth, globalConfig.Config.PageHeight))
		w.Show()
		w.SetOnClosed(func() {
			metadataPath := filepath.Dir(manga.CoverPath)
			tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer", metadataPath, manga.Title))
			err := os.RemoveAll(tmpDir)
			if err != nil {
				msg := errors.New(fmt.Sprintf("Error when trying to remove temporary view folder: %s", err))
				dialog.ShowError(msg, nil)
			}
		})
	} else {
		err := errors.New(fmt.Sprintf("The chapter %03.1f for the \"%s\" does not exists on the file system, sorry we can't read it!", chapter, manga.Name))
		dialog.ShowError(err, nil)
	}
}

func widgetReader(app fyne.App, manga settings.Manga, chapter float64) *fyne.Container {
	app.Preferences().SetFloat(manga.Title, chapter)
	metadataPath := filepath.Dir(manga.CoverPath)
	tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer", metadataPath, manga.Title))
	cbzPath := filepath.FromSlash(fmt.Sprintf("%s/%s-%03.1f.cbz", manga.Path, manga.Title, chapter))
	pages, err := archive.Unzip(cbzPath, tmpDir)
	if err != nil {
		msg := errors.New(fmt.Sprintf("Error when trying to unzip %s to temporary view folder %s: %s", cbzPath, tmpDir, err))
		dialog.ShowError(msg, nil)
	}
	pageNumber := app.Preferences().Int(fmt.Sprintf("%s/%.1f/page/", manga.Title, chapter))
	if pageNumber <= 0 {
		pageNumber = 1
	}
	//pageNumber := 1
	nbPages := len(pages)

	displayPage := widget.NewLabel(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
	pageProgress := widget.NewProgressBar()
	pageProgress.SetValue(float64(pageNumber) / float64(nbPages))

	pageView := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	pageView.File = pages[pageNumber-1]
	canvas.Refresh(pageView)

	navBar := fyne.NewContainerWithLayout(
		layout.NewGridLayout(2),
		container.NewHBox(
			widget.NewButtonWithIcon("[Prev]", theme.MediaFastRewindIcon(), func() {
				pageNumber--
				if pageNumber < 1 {
					pageNumber = 1
				}
				app.Preferences().SetInt(fmt.Sprintf("%s/%.1f/page/", manga.Title, chapter), pageNumber)
				displayPage.SetText(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
				pageProgress.SetValue(float64(pageNumber) / float64(nbPages))
				pageView.File = pages[pageNumber-1]
				canvas.Refresh(pageView)
				app.Preferences().SetInt(fmt.Sprintf("%s/page/", manga.Title), pageNumber)
			}),
			widget.NewButtonWithIcon("[Next]", theme.MediaFastForwardIcon(), func() {
				pageNumber++
				if pageNumber > nbPages {
					pageNumber = nbPages
				}
				app.Preferences().SetInt(fmt.Sprintf("%s/%.1f/page/", manga.Title, chapter), pageNumber)
				displayPage.SetText(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
				pageProgress.SetValue(float64(pageNumber) / float64(nbPages))
				pageView.File = pages[pageNumber-1]
				canvas.Refresh(pageView)
				app.Preferences().SetInt(fmt.Sprintf("%s/page/", manga.Title), pageNumber)
			}),
			layout.NewSpacer(),
			displayPage,
		),
		pageProgress,
	)

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, navBar, nil, nil), navBar, pageView)
}
