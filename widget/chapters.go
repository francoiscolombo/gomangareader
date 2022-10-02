package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"path/filepath"
)

type Chapters struct {
	widget.BaseWidget
	Manga               *settings.Manga
	Title               string
	ThumbnailPath       string
	Chapters            []string
	CurrentChapterIndex int
}

func NewChapters(manga *settings.Manga) *Chapters {
	var chaps []string
	currentChapterIndex := application.Preferences().Int(manga.Title)
	if len(manga.Chapters) == 0 {
		currentChapterIndex = 0
	}
	for i := 0; i < len(manga.Chapters); i++ {
		chaps = append(chaps, fmt.Sprintf("%03.1f", manga.Chapters[i]))
	}
	thumbnailPath := filepath.Dir(manga.CoverPath)
	nc := &Chapters{
		Manga:               manga,
		Title:               manga.Title,
		ThumbnailPath:       thumbnailPath,
		Chapters:            chaps,
		CurrentChapterIndex: currentChapterIndex,
	}
	nc.ExtendBaseWidget(nc)
	return nc
}

// MinSize returns the size that this widget should not shrink below
func (c *Chapters) MinSize() fyne.Size {
	c.ExtendBaseWidget(c)
	return c.BaseWidget.MinSize()
}

func (c *Chapters) CreateRenderer() fyne.WidgetRenderer {
	c.ExtendBaseWidget(c)

	bg := canvas.NewRectangle(theme.ButtonColor())

	thumbnail := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	thumbnail.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%s.jpg", c.ThumbnailPath, c.Title, c.Chapters[c.CurrentChapterIndex]))

	chapter := canvas.NewText(fmt.Sprintf("%s - Chapter %s", c.Title, c.Chapters[c.CurrentChapterIndex]), theme.TextColor())
	chapter.TextSize = 14

	previous := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		currentChapterIndex := c.CurrentChapterIndex - 1
		if currentChapterIndex < 1 {
			currentChapterIndex = 0
		}
		c.CurrentChapterIndex = currentChapterIndex
		c.Refresh()
	})

	next := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		currentChapterIndex := c.CurrentChapterIndex + 1
		if currentChapterIndex > (len(c.Chapters) - 1) {
			currentChapterIndex = len(c.Chapters) - 1
		}
		c.CurrentChapterIndex = currentChapterIndex
		c.Refresh()
	})

	readThis := widget.NewButtonWithIcon("Read this chapter...", theme.DocumentIcon(), func() {
		reader = NewReader(c.Manga, c.Manga.Chapters[c.CurrentChapterIndex])
		reader.Refresh()
		application.Preferences().SetInt(c.Manga.Title, c.CurrentChapterIndex)
		refreshTabsContent(c.Manga, 2)
	})

	cr := &ChaptersRenderer{
		thumbnail: thumbnail,
		previous:  previous,
		next:      next,
		readThis:  readThis,
		bg:        bg,
		chapter:   chapter,
		layout:    nil,
		chapters:  c,
	}

	return cr
}

type ChaptersRenderer struct {
	thumbnail *canvas.Image
	previous  *widget.Button
	next      *widget.Button
	readThis  *widget.Button
	bg        *canvas.Rectangle
	chapter   *canvas.Text
	layout    fyne.Layout
	chapters  *Chapters
}

func (c *ChaptersRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (c *ChaptersRenderer) Destroy() {
	c.thumbnail = nil
	c.previous = nil
	c.next = nil
	c.readThis = nil
	c.bg = nil
	c.chapter = nil
	c.chapters = nil
}

func (c *ChaptersRenderer) MinSize() fyne.Size {
	return fyne.NewSize(config.Config.ThumbMiniWidth+config.Config.LeftRightButtonWidth*2+config.Config.ChapterLabelWidth+theme.Padding()*7+200, config.Config.ThumbMiniHeight+theme.Padding()*2)
}

func (c *ChaptersRenderer) Layout(size fyne.Size) {
	p := theme.Padding()
	dx := p
	dy := p

	c.previous.Resize(fyne.NewSize(config.Config.LeftRightButtonWidth, config.Config.ThumbMiniHeight))
	c.previous.Move(fyne.NewPos(dx, dy))
	dx = dx + config.Config.LeftRightButtonWidth + p

	c.thumbnail.Resize(fyne.NewSize(config.Config.ThumbMiniWidth, config.Config.ThumbMiniHeight))
	c.thumbnail.Move(fyne.NewPos(dx, dy))
	dx = dx + config.Config.ThumbMiniWidth + p

	c.next.Resize(fyne.NewSize(config.Config.LeftRightButtonWidth, config.Config.ThumbMiniHeight))
	c.next.Move(fyne.NewPos(dx, dy))
	dx = dx + config.Config.LeftRightButtonWidth + p

	c.chapter.Resize(fyne.NewSize(config.Config.ChapterLabelWidth, config.Config.ThumbMiniHeight))
	c.chapter.Move(fyne.NewPos(dx, dy))
	dx = dx + config.Config.ChapterLabelWidth + p

	c.readThis.Resize(fyne.NewSize(200, config.Config.ThumbMiniHeight/2-p))
	c.readThis.Move(fyne.NewPos(dx, dy))
	dy = dy + config.Config.ThumbMiniHeight/2 + p

}

func (c *ChaptersRenderer) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	objects = append(objects, c.previous)
	objects = append(objects, c.thumbnail)
	objects = append(objects, c.next)
	objects = append(objects, c.chapter)
	objects = append(objects, c.readThis)
	return objects
}

func (c *ChaptersRenderer) Refresh() {

	c.thumbnail = &canvas.Image{FillMode: canvas.ImageFillOriginal}
	c.thumbnail.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%s.jpg", c.chapters.ThumbnailPath, c.chapters.Title, c.chapters.Chapters[c.chapters.CurrentChapterIndex]))
	c.thumbnail.Refresh()

	c.chapter = canvas.NewText(fmt.Sprintf("%s - Chapter %s", c.chapters.Title, c.chapters.Chapters[c.chapters.CurrentChapterIndex]), theme.TextColor())
	c.chapter.TextSize = 14
	c.chapter.Refresh()

}
