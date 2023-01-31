package widget

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
)

type Series struct {
	widget.BaseWidget
	SelectedManga *settings.Manga
}

func NewSeries(manga *settings.Manga) *Series {
	ns := &Series{
		SelectedManga: manga,
	}
	ns.ExtendBaseWidget(ns)
	return ns
}

// MinSize returns the size that this widget should not shrink below
func (s *Series) MinSize() fyne.Size {
	s.ExtendBaseWidget(s)
	return s.BaseWidget.MinSize()
}

func (s *Series) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)

	var cover *canvas.Image
	if s.SelectedManga.CoverPath != "" {
		cover = canvas.NewImageFromFile(s.SelectedManga.CoverPath)
		cover.FillMode = canvas.ImageFillContain
	}

	lName := canvas.NewText("Name:", theme.ForegroundColor())
	lName.TextSize = 12

	lAlternateName := canvas.NewText("Alternate Name:", theme.ForegroundColor())
	lAlternateName.TextSize = 12

	lStatus := canvas.NewText("Status:", theme.ForegroundColor())
	lStatus.TextSize = 12

	lNbOfChapters := canvas.NewText("Nb of chapters:", theme.ForegroundColor())
	lNbOfChapters.TextSize = 12

	lAuthor := canvas.NewText("Author:", theme.ForegroundColor())
	lAuthor.TextSize = 12

	lAvailability := canvas.NewText("Availability:", theme.ForegroundColor())
	lAvailability.TextSize = 12

	txtName := canvas.NewText(s.SelectedManga.Name, theme.ForegroundColor())
	txtName.TextSize = 12

	txtAlternateName := canvas.NewText(s.SelectedManga.AlternateName, theme.ForegroundColor())
	txtAlternateName.TextSize = 12

	txtStatus := canvas.NewText(s.SelectedManga.Status, theme.ForegroundColor())
	txtStatus.TextSize = 12

	txtNbOfChapters := canvas.NewText(fmt.Sprintf("%d", len(s.SelectedManga.Chapters)), theme.ForegroundColor())
	txtNbOfChapters.TextSize = 12

	txtAuthor := canvas.NewText(s.SelectedManga.Author, theme.ForegroundColor())
	txtAuthor.TextSize = 12

	updated := checkNewChapters(s.SelectedManga)
	txtAvailability := canvas.NewText("complete", color.NRGBA{R: 0x80, G: 0x80, B: 0xff, A: 0xff})
	if updated {
		txtAvailability = canvas.NewText("new chapters available", color.NRGBA{R: 0xff, G: 0x80, B: 0x80, A: 0xff})
	}
	txtAvailability.TextSize = 12

	txtDescription := widget.NewLabel(s.SelectedManga.Description)
	txtDescription.Wrapping = fyne.TextWrapWord
	txtDescription.TextStyle = fyne.TextStyle{
		Bold:      false,
		Italic:    true,
		Monospace: false,
	}

	bg := canvas.NewRectangle(theme.ButtonColor())

	sr := &SeriesRenderer{
		cover:          cover,
		lname:          lName,
		name:           txtName,
		lalternateName: lAlternateName,
		alternateName:  txtAlternateName,
		lstatus:        lStatus,
		status:         txtStatus,
		lnbOfChapters:  lNbOfChapters,
		nbOfChapters:   txtNbOfChapters,
		lauthor:        lAuthor,
		author:         txtAuthor,
		lavailability:  lAvailability,
		availability:   txtAvailability,
		description:    container.NewScroll(txtDescription),
		bg:             bg,
		layout:         nil,
		series:         s,
	}

	return sr
}

type SeriesRenderer struct {
	cover          *canvas.Image
	lname          *canvas.Text
	lalternateName *canvas.Text
	lstatus        *canvas.Text
	lnbOfChapters  *canvas.Text
	lauthor        *canvas.Text
	lavailability  *canvas.Text
	name           *canvas.Text
	alternateName  *canvas.Text
	status         *canvas.Text
	nbOfChapters   *canvas.Text
	author         *canvas.Text
	availability   *canvas.Text
	description    *container.Scroll
	bg             *canvas.Rectangle
	layout         fyne.Layout
	series         *Series
}

func (s *SeriesRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (s *SeriesRenderer) Destroy() {
	s.cover = nil
	s.name = nil
	s.alternateName = nil
	s.status = nil
	s.nbOfChapters = nil
	s.author = nil
	s.availability = nil
	s.description = nil
	s.lname = nil
	s.lalternateName = nil
	s.lstatus = nil
	s.lnbOfChapters = nil
	s.lauthor = nil
	s.lavailability = nil
	s.bg = nil
	s.layout = nil
	s.series = nil
}

func (s *SeriesRenderer) Layout(_ fyne.Size) {
	var txtHeight float32
	p := theme.Padding()
	ldx := (config.Config.ThumbnailWidth + p) * 2
	dx := (config.Config.ThumbnailWidth + p) * 3
	dy := p
	txtHeight = 20.0

	s.cover.Resize(fyne.NewSize(config.Config.ThumbnailWidth*2, config.Config.ThumbnailHeight*2))
	s.cover.Move(fyne.NewPos(p, p))

	s.lname.Move(fyne.NewPos(ldx, dy))
	s.name.Move(fyne.NewPos(dx, dy))
	dy = dy + txtHeight + p

	s.lalternateName.Move(fyne.NewPos(ldx, dy))
	s.alternateName.Move(fyne.NewPos(dx, dy))
	dy = dy + txtHeight + p

	s.lstatus.Move(fyne.NewPos(ldx, dy))
	s.status.Move(fyne.NewPos(dx, dy))
	dy = dy + txtHeight + p

	s.lnbOfChapters.Move(fyne.NewPos(ldx, dy))
	s.nbOfChapters.Move(fyne.NewPos(dx, dy))
	dy = dy + txtHeight + p

	s.lauthor.Move(fyne.NewPos(ldx, dy))
	s.author.Move(fyne.NewPos(dx, dy))
	dy = dy + txtHeight + p

	s.lavailability.Move(fyne.NewPos(ldx, dy))
	s.availability.Move(fyne.NewPos(dx, dy))
	dy = dy + txtHeight + p

	s.description.Resize(fyne.NewSize(config.Config.ThumbnailWidth*4, config.Config.ThumbTextHeight*12))
	s.description.Move(fyne.NewPos(ldx, dy))

}

func (s *SeriesRenderer) MinSize() fyne.Size {
	return fyne.NewSize(config.Config.ThumbnailWidth*6+theme.Padding()*2, config.Config.ThumbnailHeight*2+theme.Padding()*2)
}

func (s *SeriesRenderer) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	objects = append(objects, s.cover)
	objects = append(objects, s.lname)
	objects = append(objects, s.name)
	objects = append(objects, s.lalternateName)
	objects = append(objects, s.alternateName)
	objects = append(objects, s.lstatus)
	objects = append(objects, s.status)
	objects = append(objects, s.lnbOfChapters)
	objects = append(objects, s.nbOfChapters)
	objects = append(objects, s.lauthor)
	objects = append(objects, s.author)
	objects = append(objects, s.lavailability)
	objects = append(objects, s.availability)
	objects = append(objects, s.description)
	return objects
}

func (s *SeriesRenderer) Refresh() {
	s.bg.Refresh()
	s.cover.Refresh()
	s.lname.Refresh()
	s.name.Refresh()
	s.lalternateName.Refresh()
	s.alternateName.Refresh()
	s.lstatus.Refresh()
	s.status.Refresh()
	s.lnbOfChapters.Refresh()
	s.nbOfChapters.Refresh()
	s.lauthor.Refresh()
	s.author.Refresh()
	s.lavailability.Refresh()
	s.availability.Refresh()
	s.description.Refresh()
}
