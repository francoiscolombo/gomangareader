package widget

import (
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/archive"
	"github.com/francoiscolombo/gomangareader/settings"
	"os"
	"path/filepath"
)

func isMangaChapterExists(manga settings.Manga, chapter int) bool {
	cbzPath := filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.cbz", manga.Path, manga.Title, chapter))
	info, err := os.Stat(cbzPath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func readChapter(app fyne.App, win fyne.Window, manga settings.Manga, chapter int) {
	if isMangaChapterExists(manga, chapter) {
		w := app.NewWindow(fmt.Sprintf("Read %s - chapter %d", manga.Name, chapter))
		w.SetContent(widget.NewScrollContainer(widgetReader(app, w, manga, chapter)))
		w.Resize(fyne.NewSize(pgWidth, pgHeight))
		w.Show()
		w.SetOnClosed(func() {
			metadataPath := filepath.Dir(manga.CoverPath)
			tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer", metadataPath, manga.Title))
			err := os.RemoveAll(tmpDir)
			if err != nil {
				msg := errors.New(fmt.Sprintf("Error when trying to remove temporary view folder: %s", err))
				dialog.ShowError(msg, win)
			}
		})
	} else {
		err := errors.New(fmt.Sprintf("The chapter %03d for the \"%s\" does not exists on the file system, sorry we can't read it!", chapter, manga.Name))
		dialog.ShowError(err, win)
	}
}

func widgetReader(app fyne.App, win fyne.Window, manga settings.Manga, chapter int) *fyne.Container {
	app.Preferences().SetInt(manga.Title, chapter)
	metadataPath := filepath.Dir(manga.CoverPath)
	tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer", metadataPath, manga.Title))
	cbzPath := filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.cbz", manga.Path, manga.Title, chapter))
	pages, err := archive.Unzip(cbzPath, tmpDir)
	if err != nil {
		msg := errors.New(fmt.Sprintf("Error when trying to unzip %s to temporary view folder %s: %s", cbzPath, tmpDir, err))
		dialog.ShowError(msg, win)
	}
	pageNumber := app.Preferences().Int(fmt.Sprintf("%s/%d/page/", manga.Title, chapter))
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
		widget.NewHBox(
			widget.NewButtonWithIcon("[Prev]", theme.MediaFastRewindIcon(), func() {
				pageNumber--
				if pageNumber < 1 {
					pageNumber = 1
				}
				app.Preferences().SetInt(fmt.Sprintf("%s/%d/page/", manga.Title, chapter), pageNumber)
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
				app.Preferences().SetInt(fmt.Sprintf("%s/%d/page/", manga.Title, chapter), pageNumber)
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
