package widget

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
)

type TitleButton struct {
	widget.BaseWidget
	Title    *settings.Manga
	Selected bool
}

func NewTitleButton(title settings.Manga) *TitleButton {
	tb := &TitleButton{
		Title:    &title,
		Selected: false,
	}
	tb.ExtendBaseWidget(tb)
	return tb
}

func (t *TitleButton) CreateRenderer() fyne.WidgetRenderer {
	t.ExtendBaseWidget(t)

	var cover *canvas.Image
	if t.Title.CoverPath != "" {
		cover = canvas.NewImageFromFile(t.Title.CoverPath)
		cover.FillMode = canvas.ImageFillContain
	}

	title := t.Title.Name
	if len(title) > 20 {
		title = title[0:17] + "..."
	}
	colorTitle := color.NRGBA{R: 0x80, G: 0xff, A: 0xff}
	if checkNewChapters(t.Title) {
		colorTitle = color.NRGBA{R: 0xff, G: 0x80, A: 0xff}
	}
	text := canvas.NewText(title, colorTitle)
	text.TextSize = 10

	bg := canvas.NewRectangle(theme.ButtonColor())

	r := &TitleButtonRenderer{
		cover:       cover,
		title:       text,
		bg:          bg,
		titleButton: t,
		layout:      layout.NewVBoxLayout(),
	}

	return r
}

// MinSize returns the size that this widget should not shrink below
func (t *TitleButton) MinSize() fyne.Size {
	t.ExtendBaseWidget(t)
	return t.BaseWidget.MinSize()
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (t *TitleButton) Tapped(*fyne.PointEvent) {
	for _, tb := range library.Items {
		tb.Selected = false
	}
	t.Selected = true
	library.Refresh()

	refreshTabsContent(t.Title, 1)

}

type TitleButtonRenderer struct {
	cover       *canvas.Image
	title       *canvas.Text
	bg          *canvas.Rectangle
	titleButton *TitleButton
	layout      fyne.Layout
}

func (t *TitleButtonRenderer) BackgroundColor() color.Color {
	return theme.ButtonColor()
}

func (t *TitleButtonRenderer) Destroy() {
	t.bg = nil
	t.cover = nil
	t.title = nil
	t.titleButton = nil
}

func (t *TitleButtonRenderer) MinSize() fyne.Size {
	return fyne.NewSize(config.Config.ThumbnailWidth, config.Config.ThumbnailHeight+config.Config.ThumbTextHeight)
}

func (t *TitleButtonRenderer) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	objects = append(objects, t.cover, t.title)
	return objects
}

func (t *TitleButtonRenderer) Layout(size fyne.Size) {
	//log.Printf(">>> method Layout with size %d/%d called on %s title button", size.Width, size.Height, t.title)
	t.bg.Resize(fyne.NewSize(config.Config.ThumbnailWidth+theme.Padding()/2, config.Config.ThumbnailHeight+config.Config.ThumbTextHeight+theme.Padding()/2))
	t.cover.SetMinSize(fyne.NewSize(config.Config.ThumbnailWidth, config.Config.ThumbnailHeight))
	t.title.SetMinSize(fyne.NewSize(config.Config.ThumbnailWidth, config.Config.ThumbTextHeight))
	objects := []fyne.CanvasObject{t.cover, t.title}
	min := t.layout.MinSize(objects)
	t.layout.Layout(objects, min)
}

func (t *TitleButtonRenderer) Refresh() {
	if t.titleButton.Selected {
		t.bg.StrokeWidth = 1.0
		t.bg.StrokeColor = color.NRGBA{R: 0x40, G: 0x80, B: 0xff, A: 0xff}
		t.bg.FillColor = color.NRGBA{R: 0x20, G: 0x60, B: 0x80, A: 0xff}
	} else {
		t.bg.FillColor = color.Transparent
	}
	t.bg.Refresh()
	t.cover.FillMode = canvas.ImageFillContain
	t.cover.Refresh()
	t.cover.Show()
	t.title.TextStyle = fyne.TextStyle{
		Bold:      t.titleButton.Selected,
		Italic:    t.titleButton.Selected,
		Monospace: false,
	}
	t.title.Alignment = fyne.TextAlignCenter
	t.title.Refresh()
	t.title.Show()
	t.Layout(t.titleButton.Size())
	canvas.Refresh(t.titleButton)
}

func checkNewChapters(manga *settings.Manga) bool {
	var provider settings.MangaProvider
	provider = settings.MangaReader{}
	checkLastChapter := provider.CheckLastChapter(*manga)
	if manga.LastChapter < checkLastChapter {
		return true
	}
	return false
}
