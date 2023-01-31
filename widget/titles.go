package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"sort"
)

func getLastChapterIndex(manga settings.Manga) int {
	index := 0
	for i, c := range manga.Chapters {
		if c == manga.LastChapter {
			index = i
			break
		}
	}
	return index
}

type Titles struct {
	widget.BaseWidget
	Items []*TitleButton
	Manga *settings.Manga
}

// NewTitlesContainer create a new titles container
func NewTitlesContainer(items ...*TitleButton) *Titles {
	var t *Titles
	if items == nil {
		t = &Titles{
			BaseWidget: widget.BaseWidget{},
			Items:      nil,
			Manga:      nil,
		}
	} else {
		t = &Titles{
			BaseWidget: widget.BaseWidget{},
			Items:      items,
			Manga:      items[0].Title,
		}
		t.Items[0].Selected = true
		series = NewSeries(items[0].Title)
		chapters = NewChapters(items[0].Title)
		downloader = NewDownloader(items[0].Title, getLastChapterIndex(*items[0].Title))
	}
	t.ExtendBaseWidget(t)
	return t
}

// Add adds the given item to this container
func (t *Titles) Add(item *TitleButton) {
	if series == nil {
		item.Selected = true
		series = NewSeries(item.Title)
	}
	if chapters == nil {
		chapters = NewChapters(item.Title)
	}
	if downloader == nil {
		downloader = NewDownloader(item.Title, getLastChapterIndex(*item.Title))
	}
	t.Items = append(t.Items, item)
	sort.Slice(t.Items, func(i, j int) bool {
		return t.Items[i].Title.Title < t.Items[j].Title.Title
	})
	t.Refresh()
}

func (t *Titles) CreateRenderer() fyne.WidgetRenderer {
	t.ExtendBaseWidget(t)
	bg := canvas.NewRectangle(theme.ButtonColor())
	r := &TitlesRenderer{
		bg:        bg,
		layout:    layout.NewGridLayout(int(config.Config.NbColumns)),
		container: t,
	}
	return r
}

// MinSize returns the size that this widget should not shrink below
func (t *Titles) MinSize() fyne.Size {
	t.ExtendBaseWidget(t)
	return t.BaseWidget.MinSize()
}

// Remove deletes the given item from this container.
func (t *Titles) Remove(item *TitleButton) {
	for i, ti := range t.Items {
		if ti == item {
			t.RemoveIndex(i)
			break
		}
	}
	t.Refresh()
}

// RemoveIndex deletes the item at the given index from this container.
func (t *Titles) RemoveIndex(index int) {
	if index < 0 || index >= len(t.Items) {
		return
	}
	t.Items = append(t.Items[:index], t.Items[index+1:]...)
}

type TitlesRenderer struct {
	bg        *canvas.Rectangle
	layout    fyne.Layout
	container *Titles
}

func (t *TitlesRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (t *TitlesRenderer) Destroy() {
	t.container.Items = nil
	t.container = nil
}

func (t *TitlesRenderer) MinSize() fyne.Size {
	var nbRows float32
	nbRows = float32((len(t.container.Items) / 6) + 1)
	return fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns, (config.Config.ThumbnailHeight+config.Config.ThumbTextHeight)*nbRows+40)
}

func (t *TitlesRenderer) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	for _, tb := range t.container.Items {
		objects = append(objects, tb)
	}
	return objects
}

func (t *TitlesRenderer) Layout(_ fyne.Size) {
	t.bg.Resize(t.MinSize())
	objects := t.Objects()
	min := t.layout.MinSize(objects)
	t.layout.Layout(objects, min)
}

func (t *TitlesRenderer) Refresh() {
	for _, tb := range t.container.Items {
		tb.Refresh()
		if tb.Selected == true {
			t.container.Manga = tb.Title
		}
	}
	t.bg.Refresh()
	t.Layout(t.container.Size())
	canvas.Refresh(t.container)
}
