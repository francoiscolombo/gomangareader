package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/container"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"path/filepath"
	"strconv"
	"strings"
)

func checkNewChapters(manga settings.Manga) bool {
	var provider settings.MangaProvider
	if manga.Provider == "mangareader.cc" {
		provider = settings.MangaReader{}
		return provider.CheckLastChapter(manga) >= manga.LastChapter
	} else if manga.Provider == "mangapanda.in" {
		provider = settings.MangaPanda{}
		return provider.CheckLastChapter(manga) >= manga.LastChapter
	}
	return false
}

func chaptersPanel(app fyne.App, win fyne.Window, manga settings.Manga, available bool) *fyne.Container {
	var chaps []string
	currentChapter := app.Preferences().Int(manga.Title)
	if currentChapter <= 0 {
		currentChapter = 1
	}
	nbChapters := manga.LastChapter - 1
	for i := 1; i <= nbChapters; i++ {
		chaps = append(chaps, fmt.Sprintf("Chapter %d / %d", i, nbChapters))
	}
	thumbnailPath := filepath.Dir(manga.CoverPath)

	thumbnailView := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	thumbnailView.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.jpg", thumbnailPath, manga.Title, currentChapter))
	canvas.Refresh(thumbnailView)

	selChapter := widget.NewSelect(chaps, func(s string) {
		f := strings.Fields(s)
		currentChapter, _ = strconv.Atoi(f[1])
		thumbnailView.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.jpg", thumbnailPath, manga.Title, currentChapter))
		canvas.Refresh(thumbnailView)
	})
	selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d", currentChapter, nbChapters))
	prev := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		currentChapter--
		if currentChapter < 1 {
			currentChapter = 1
		}
		selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d", currentChapter, nbChapters))
	})
	next := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		currentChapter++
		if currentChapter > nbChapters {
			currentChapter = nbChapters
		}
		selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d", currentChapter, nbChapters))
	})
	read := widget.NewButtonWithIcon("Read this chapter...", theme.DocumentIcon(), func() {
		readChapter(app, win, manga, currentChapter)
	})
	if available {
		read.Text = "Read this..."
		download := widget.NewButtonWithIcon("Download chapters...", theme.DocumentSaveIcon(), func() {
			showDownloadChapters(app, win, manga)
		})
		return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2, chHeight)), prev),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth, chHeight)), thumbnailView),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2, chHeight)), next),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*1.5, chHeight)), selChapter),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, chHeight)), read),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*1.5, chHeight)), download))
	}
	return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2, chHeight)), prev),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth, chHeight)), thumbnailView),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2, chHeight)), next),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2, chHeight)), selChapter),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2, chHeight)), read))
}

func widgetDetailSerie(app fyne.App, win fyne.Window, manga settings.Manga) *fyne.Container {
	coverImage := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2, thHeight*2)), canvas.NewImageFromFile(manga.CoverPath))
	updated := checkNewChapters(manga)
	nca := canvas.NewText("complete", color.NRGBA{R: 0x80, G: 0x80, B: 0xff, A: 0xff})
	if updated {
		nca = canvas.NewText("new chapters available", color.NRGBA{R: 0xff, G: 0x80, B: 0x80, A: 0xff})
	}
	availability := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2, 30)), nca)
	detailPanel := fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, 30)),
			widget.NewLabel("Provider:"),
			widget.NewLabel("Name:"),
			widget.NewLabel("Alternate Name:"),
			widget.NewLabel("Year of release:"),
			widget.NewLabel("Status:"),
			widget.NewLabel("Nb of Chapters:"),
			widget.NewLabel("Author:"),
			widget.NewLabel("Artist:"),
			availability),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, 30)),
			widget.NewLabel(manga.Provider),
			widget.NewLabel(manga.Name),
			widget.NewLabel(manga.AlternateName),
			widget.NewLabel(manga.YearOfRelease),
			widget.NewLabel(manga.Status),
			widget.NewLabel(fmt.Sprintf("%d", manga.LastChapter-1)),
			widget.NewLabel(manga.Author),
			widget.NewLabel(manga.Artist),
			widget.NewLabel("")))
	desc := widget.NewLabel(manga.Description)
	desc.Wrapping = fyne.TextWrapWord
	description := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5, thHeight)), container.NewScroll(desc))
	browseChapters := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5, chHeight)), chaptersPanel(app, win, manga, updated))
	return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewVBoxLayout(), fyne.NewContainerWithLayout(layout.NewHBoxLayout(), coverImage, detailPanel), browseChapters, description))
}

func showSerieDetail(app fyne.App, manga settings.Manga) {
	w := app.NewWindow(fmt.Sprintf("%s - details", manga.Name))
	w.SetContent(widgetDetailSerie(app, w, manga))
	w.Resize(fyne.NewSize(thWidth*3, thHeight*2))
	w.Show()
}
