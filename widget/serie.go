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
		checkLastChapter := provider.CheckLastChapter(manga)
		if manga.LastChapter < checkLastChapter {
			return true
		}
	}
	return false
}

func chaptersPanel(app fyne.App, win fyne.Window, manga settings.Manga, available bool) *fyne.Container {
	var chaps []string
	currentChapterIndex := app.Preferences().Int(manga.Title)
	currentChapter := 0.0
	if len(manga.Chapters) == 0 {
		currentChapter = 1.0
	} else {
		currentChapter = manga.Chapters[currentChapterIndex]
	}

	for i := 0; i < len(manga.Chapters); i++ {
		chaps = append(chaps, fmt.Sprintf("Chapter %.1f", manga.Chapters[i]))
	}
	thumbnailPath := filepath.Dir(manga.CoverPath)

	thumbnailView := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	thumbnailView.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%03.1f.jpg", thumbnailPath, manga.Title, currentChapter))
	canvas.Refresh(thumbnailView)

	selChapter := widget.NewSelect(chaps, func(s string) {
		f := strings.Fields(s)
		currentChapter, _ = strconv.ParseFloat(f[1], 64)
		for i := 0; i < len(manga.Chapters); i++ {
			if fmt.Sprintf("3.1f", manga.Chapters[i]) == fmt.Sprintf("3.1f", currentChapter) {
				currentChapterIndex = i
				break
			}
		}
		thumbnailView.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%03.1f.jpg", thumbnailPath, manga.Title, currentChapter))
		canvas.Refresh(thumbnailView)
	})
	selChapter.SetSelected(fmt.Sprintf("Chapter %.1f", currentChapter))
	prev := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		currentChapterIndex = currentChapterIndex - 1
		if currentChapterIndex < 1 {
			currentChapterIndex = 0
		}
		currentChapter = manga.Chapters[currentChapterIndex]
		selChapter.SetSelected(fmt.Sprintf("Chapter %.1f", currentChapter))
	})
	next := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		currentChapterIndex = currentChapterIndex + 1
		if currentChapterIndex > (len(manga.Chapters) - 1) {
			currentChapterIndex = (len(manga.Chapters) - 1)
		}
		currentChapter = manga.Chapters[currentChapterIndex]
		selChapter.SetSelected(fmt.Sprintf("Chapter %.1f", currentChapter))
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
			widget.NewLabel(fmt.Sprintf("%d", len(manga.Chapters)+1)),
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
