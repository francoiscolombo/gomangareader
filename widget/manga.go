package widget

import (
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/archive"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"os"
	"path/filepath"
)

type Reader struct {
	widget.BaseWidget
	Manga      *settings.Manga
	Chapter    float64
	PageNumber int
	NbPages    int
	TempDir    string
	Pages      []string
}

func NewReader(manga *settings.Manga, chapter float64) *Reader {

	application.Preferences().SetFloat(manga.Title, chapter)

	metadataPath := filepath.Dir(manga.CoverPath)
	tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer", metadataPath, manga.Title))
	cbzPath := filepath.FromSlash(fmt.Sprintf("%s/%s-%03.1f.cbz", manga.Path, manga.Title, chapter))
	pages, err := archive.Unzip(cbzPath, tmpDir)
	if err != nil {
		msg := errors.New(fmt.Sprintf("Error when trying to unzip %s to temporary view folder %s: %s", cbzPath, tmpDir, err))
		dialog.ShowError(msg, mainWindow)
	}
	pageNumber := application.Preferences().Int(fmt.Sprintf("%s/%.1f/page/", manga.Title, chapter))
	if pageNumber <= 0 {
		pageNumber = 1
	}
	nbPages := len(pages)

	nr := &Reader{
		Manga:      manga,
		Chapter:    chapter,
		PageNumber: pageNumber,
		NbPages:    nbPages,
		Pages:      pages,
		TempDir:    tmpDir,
	}
	nr.ExtendBaseWidget(nr)
	return nr
}

// MinSize returns the size that this widget should not shrink below
func (r *Reader) MinSize() fyne.Size {
	r.ExtendBaseWidget(r)
	return r.BaseWidget.MinSize()
}

func (r *Reader) CreateRenderer() fyne.WidgetRenderer {
	r.ExtendBaseWidget(r)

	bg := canvas.NewRectangle(theme.ButtonColor())

	displayPage := canvas.NewText(fmt.Sprintf("Page %d / %d", r.PageNumber, r.NbPages), theme.TextColor())
	displayPage.TextStyle = fyne.TextStyle{Monospace: true}
	displayPage.TextSize = 10

	pageProgress := widget.NewProgressBar()
	pageProgress.SetValue(float64(r.PageNumber) / float64(r.NbPages))

	pageView := &canvas.Image{FillMode: canvas.ImageFillContain}
	pageView.File = r.Pages[r.PageNumber-1]

	prev := widget.NewButtonWithIcon("[Prev]", theme.MediaFastRewindIcon(), func() {
		r.PageNumber--
		if r.PageNumber < 1 {
			r.PageNumber = 1
		}
		r.Refresh()
	})

	next := widget.NewButtonWithIcon("[Next]", theme.MediaFastForwardIcon(), func() {
		r.PageNumber++
		if r.PageNumber > r.NbPages {
			r.PageNumber = r.NbPages
		}
		r.Refresh()
	})

	rr := &ReaderRenderer{
		bg:           bg,
		page:         pageView,
		displayPage:  displayPage,
		pageProgress: pageProgress,
		previous:     prev,
		next:         next,
		layout:       nil,
		reader:       r,
	}

	return rr
}

type ReaderRenderer struct {
	bg           *canvas.Rectangle
	page         *canvas.Image
	displayPage  *canvas.Text
	pageProgress *widget.ProgressBar
	previous     *widget.Button
	next         *widget.Button
	layout       fyne.Layout
	reader       *Reader
}

func (r *ReaderRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (r *ReaderRenderer) MinSize() fyne.Size {
	return fyne.NewSize(libraryTabs.Size().Width, libraryTabs.Size().Height-r.previous.MinSize().Height)
}

func (r *ReaderRenderer) Destroy() {
	err := os.RemoveAll(r.reader.TempDir)
	if err != nil {
		msg := errors.New(fmt.Sprintf("Error when trying to remove temporary view folder: %s", err))
		dialog.ShowError(msg, mainWindow)
	}
	r.bg = nil
	r.page = nil
	r.displayPage = nil
	r.pageProgress = nil
	r.previous = nil
	r.next = nil
	r.layout = nil
	r.reader = nil
}

func (r *ReaderRenderer) Layout(size fyne.Size) {
	p := theme.Padding()

	dx := p
	dy := p

	r.page.Resize(fyne.NewSize(libraryTabs.Size().Width-p*2, libraryTabs.Size().Height-p*3-r.previous.MinSize().Height*2))
	r.page.Move(fyne.NewPos(dx, dy))
	dy = dy + libraryTabs.Size().Height - p*2 - r.previous.MinSize().Height*2

	//r.displayPage.Resize(r.displayPage.MinSize())
	//r.displayPage.Move(fyne.NewPos(dx, dy))
	//dx = dx + p + r.displayPage.MinSize().Width

	r.previous.Resize(r.previous.MinSize())
	r.previous.Move(fyne.NewPos(dx, dy))
	dx = dx + p + r.previous.MinSize().Width

	r.next.Resize(r.next.MinSize())
	r.next.Move(fyne.NewPos(dx, dy))
	dx = dx + p + r.next.MinSize().Width

	r.pageProgress.Resize(fyne.NewSize(libraryTabs.Size().Width-p-dx, r.next.MinSize().Height))
	r.pageProgress.Move(fyne.NewPos(dx, dy))
}

func (r *ReaderRenderer) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	objects = append(objects, r.bg)
	objects = append(objects, r.page)
	objects = append(objects, r.previous)
	//objects = append(objects, r.displayPage)
	objects = append(objects, r.next)
	objects = append(objects, r.pageProgress)
	return objects
}

func (r *ReaderRenderer) Refresh() {
	//r.displayPage = canvas.NewText(fmt.Sprintf("Page %d / %d", r.reader.PageNumber, r.reader.NbPages), theme.TextColor())
	//r.displayPage.Refresh()

	r.pageProgress.SetValue(float64(r.reader.PageNumber) / float64(r.reader.NbPages))
	r.pageProgress.Refresh()

	r.page = &canvas.Image{FillMode: canvas.ImageFillContain}
	r.page.File = r.reader.Pages[r.reader.PageNumber-1]
	r.page.Refresh()
}
