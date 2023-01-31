package widget

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"sort"
)

type SearchItem struct {
	widget.BaseWidget
	MangaFound *settings.Manga
	isSelected bool
}

func NewSearchItem(manga settings.Manga) *SearchItem {
	nsi := &SearchItem{
		MangaFound: &manga,
		isSelected: checkMangaAlreadyInLibrary(manga),
	}
	nsi.ExtendBaseWidget(nsi)
	return nsi
}

// MinSize returns the size that this widget should not shrink below
func (si *SearchItem) MinSize() fyne.Size {
	si.ExtendBaseWidget(si)
	return si.BaseWidget.MinSize()
}

// Tapped is called when a pointer tapped event is captured and triggers any tap handler
func (si *SearchItem) Tapped(*fyne.PointEvent) {
	if si.isSelected == false {
		confirm := dialog.NewConfirm(
			"Add to the library?",
			fmt.Sprintf("Do you really want to add\n%s\nto you library now?", si.MangaFound.Name),
			func(selected bool) {
				if selected == true {
					si.isSelected = true
					si.Refresh()
					provider := settings.MangaReader{}
					newManga := provider.FindDetails(config.Config.LibraryPath, si.MangaFound.Title, 0)
					provider.BuildChaptersList(&newManga)
					// download cover picture (if needed)
					err1 := downloadCover(newManga)
					if err1 == nil {
						// and generate thumbnails (if needed)
						err2 := extractFirstPages(config.Config.LibraryPath, newManga)
						if err2 == nil {
							// update library
							ws := NewTitleButton(newManga)
							library.Add(ws)
							library.Refresh()
							// and update history
							config.History.Titles = append(config.History.Titles, newManga)
							sort.Slice(config.History.Titles, func(i, j int) bool {
								return config.History.Titles[i].Title < config.History.Titles[j].Title
							})
							// okay we have updated the metadata, now we can save the config
							newSettings := settings.Settings{
								Config: settings.Config{
									LibraryPath:          config.Config.LibraryPath,
									AutoUpdate:           config.Config.AutoUpdate,
									NbColumns:            config.Config.NbColumns,
									NbRows:               config.Config.NbRows,
									PageWidth:            config.Config.PageWidth,
									PageHeight:           config.Config.PageHeight,
									ThumbMiniWidth:       config.Config.ThumbMiniWidth,
									ThumbMiniHeight:      config.Config.ThumbMiniHeight,
									LeftRightButtonWidth: config.Config.LeftRightButtonWidth,
									ChapterLabelWidth:    config.Config.ChapterLabelWidth,
									ThumbnailWidth:       config.Config.ThumbnailWidth,
									ThumbnailHeight:      config.Config.ThumbnailHeight,
									ThumbTextHeight:      config.Config.ThumbTextHeight,
									NbWorkers:            config.Config.NbWorkers,
								},
								History: settings.History{
									Titles: config.History.Titles,
								},
							}
							settings.WriteSettings(newSettings)
						}
					}
				}
			},
			mainWindow,
		)
		confirm.SetConfirmText("Of course, I want!")
		confirm.SetDismissText("No, I changed my mind.")
		confirm.Show()
	}
}

func (si *SearchItem) CreateRenderer() fyne.WidgetRenderer {
	si.ExtendBaseWidget(si)

	bg := canvas.NewRectangle(theme.ButtonColor())

	thumbnail := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	if si.isSelected {
		thumbnail.Resource = theme.FolderOpenIcon()
	} else {
		thumbnail.Resource = theme.FolderNewIcon()
	}

	lineUp := &canvas.Line{
		Hidden:      false,
		StrokeColor: color.White,
		StrokeWidth: 1,
	}

	lineDown := &canvas.Line{
		Hidden:      false,
		StrokeColor: color.White,
		StrokeWidth: 1,
	}

	title := canvas.NewText(si.MangaFound.Name, theme.ForegroundColor())
	title.TextStyle = fyne.TextStyle{Italic: true}

	description := widget.NewLabel("")
	if len(si.MangaFound.Description) > 250 {
		description.Text = si.MangaFound.Description[:250] + "..."
	} else {
		description.Text = si.MangaFound.Description
	}
	description.Wrapping = fyne.TextWrapWord

	sir := &SearchItemRender{
		bg:          bg,
		thumbnail:   thumbnail,
		lineUp:      lineUp,
		title:       title,
		description: description,
		lineDown:    lineDown,
		layout:      nil,
		item:        si,
	}

	return sir
}

type SearchItemRender struct {
	bg          *canvas.Rectangle
	thumbnail   *canvas.Image
	lineUp      *canvas.Line
	title       *canvas.Text
	description *widget.Label
	lineDown    *canvas.Line
	layout      fyne.Layout
	item        *SearchItem
}

func (s *SearchItemRender) BackgroundColor() color.Color {
	return theme.HoverColor()
}

func (s *SearchItemRender) Destroy() {
	s.bg = nil
	s.title = nil
	s.description = nil
	s.layout = nil
	s.lineDown = nil
	s.lineUp = nil
	s.thumbnail = nil
	s.item = nil
}

func (s *SearchItemRender) MinSize() fyne.Size {
	return fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns, config.Config.ThumbMiniHeight+theme.Padding()*2)
}

func (s *SearchItemRender) Layout(_ fyne.Size) {
	p := theme.Padding()

	s.lineUp.Resize(fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns, 1))
	s.lineUp.Move(fyne.NewPos(0, 1))

	s.lineDown.Resize(fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns, 1))
	s.lineDown.Move(fyne.NewPos(0, config.Config.ThumbMiniHeight+p*2))

	dx := p
	dy := p

	s.thumbnail.Resize(fyne.NewSize(config.Config.ThumbMiniWidth, config.Config.ThumbMiniHeight))
	s.thumbnail.Move(fyne.NewPos(dx, dy))
	dx = dx + p + config.Config.ThumbMiniWidth

	s.title.Resize(fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns-p-dx, config.Config.ThumbTextHeight))
	s.title.Move(fyne.NewPos(dx, dy))
	dy = dy + config.Config.ThumbTextHeight + p

	s.description.Resize(fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns-p-dx, config.Config.ThumbMiniHeight))
	s.description.Move(fyne.NewPos(dx, dy))

}

func (s *SearchItemRender) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	objects = append(objects, s.bg)
	objects = append(objects, s.lineUp)
	objects = append(objects, s.thumbnail)
	objects = append(objects, s.title)
	objects = append(objects, s.description)
	objects = append(objects, s.lineDown)
	return objects
}

func (s *SearchItemRender) Refresh() {
	if s.item.isSelected {
		s.thumbnail.Resource = theme.FolderOpenIcon()
	} else {
		s.thumbnail.Resource = theme.FolderNewIcon()
	}
	s.thumbnail.Refresh()
}

func checkMangaAlreadyInLibrary(manga settings.Manga) bool {
	for _, m := range config.History.Titles {
		if m.Title == manga.Title {
			return true
		}
	}
	return false
}
