package widget

import (
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
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
	Library    fyne.App
	Items      []*TitleButton
	Manga      *settings.Manga
	Series     *Series
	Chapters   *Chapters
	Downloader *Downloader
}

// NewTitlesContainer create a new titles container
func NewTitlesContainer(library fyne.App, items ...*TitleButton) *Titles {
	var t *Titles
	if items == nil {
		t = &Titles{
			BaseWidget: widget.BaseWidget{},
			Library:    library,
			Items:      nil,
			Manga:      nil,
			Series:     nil,
			Chapters:   nil,
			Downloader: nil,
		}
	} else {
		t = &Titles{
			BaseWidget: widget.BaseWidget{},
			Library:    library,
			Items:      items,
			Manga:      items[0].Title,
			Series:     NewSeries(items[0].Title),
			Chapters:   NewChapters(library, items[0].Title),
			Downloader: NewDownloader(library, items[0].Title, getLastChapterIndex(*items[0].Title)),
		}
		t.Items[0].Selected = true
	}
	t.ExtendBaseWidget(t)
	return t
}

// Add adds the given item to this container
func (t *Titles) Add(item *TitleButton) {
	if t.Series == nil {
		item.Selected = true
		t.Series = NewSeries(item.Title)
	}
	if t.Chapters == nil {
		t.Chapters = NewChapters(t.Library, item.Title)
	}
	if t.Downloader == nil {
		t.Downloader = NewDownloader(t.Library, item.Title, getLastChapterIndex(*item.Title))
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
		layout:    layout.NewGridLayout(globalConfig.Config.NbColumns),
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
	nbRows := (len(t.container.Items) / 6) + 1
	return fyne.NewSize(globalConfig.Config.ThumbnailWidth*globalConfig.Config.NbColumns, (globalConfig.Config.ThumbnailHeight+globalConfig.Config.ThumbTextHeight)*nbRows+40)
}

func (t *TitlesRenderer) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	for _, tb := range t.container.Items {
		objects = append(objects, tb)
	}
	return objects
}

func (t *TitlesRenderer) Layout(size fyne.Size) {
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
			t.container.Series = NewSeries(tb.Title)
			t.container.Chapters = NewChapters(tb.Library, tb.Title)
			t.container.Downloader = NewDownloader(tb.Library, tb.Title, getLastChapterIndex(*tb.Title))
		}
	}
	t.bg.Refresh()
	t.Layout(t.container.Size())
	canvas.Refresh(t.container)
}
