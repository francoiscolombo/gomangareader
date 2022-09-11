package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"sort"
)

type SearchItem struct {
	widget.BaseWidget
	Application fyne.App
	MangaFound  *settings.Manga
	isSelected  bool
}

func NewSearchItem(app fyne.App, manga settings.Manga) *SearchItem {
	nsi := &SearchItem{
		Application: app,
		MangaFound:  &manga,
		isSelected:  checkMangaAlreadyInLibrary(manga),
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
					newManga := provider.FindDetails(globalConfig.Config.LibraryPath, si.MangaFound.Title, 0)
					provider.BuildChaptersList(&newManga)
					// download cover picture (if needed)
					err1 := downloadCover(newManga)
					if err1 == nil {
						// and generate thumbnails (if needed)
						err2 := extractFirstPages(globalConfig.Config.LibraryPath, newManga)
						if err2 == nil {
							// update library
							ws := NewTitleButton(si.Application, titles, newManga)
							titles.Add(ws)
							titles.Refresh()
							// and update history
							globalConfig.History.Titles = append(globalConfig.History.Titles, newManga)
							sort.Slice(globalConfig.History.Titles, func(i, j int) bool {
								return globalConfig.History.Titles[i].Title < globalConfig.History.Titles[j].Title
							})
							// okay we have updated the metadata, now we can save the config
							newSettings := settings.Settings{
								Config: settings.Config{
									LibraryPath:          globalConfig.Config.LibraryPath,
									AutoUpdate:           globalConfig.Config.AutoUpdate,
									NbColumns:            globalConfig.Config.NbColumns,
									NbRows:               globalConfig.Config.NbRows,
									PageWidth:            globalConfig.Config.PageWidth,
									PageHeight:           globalConfig.Config.PageHeight,
									ThumbMiniWidth:       globalConfig.Config.ThumbMiniWidth,
									ThumbMiniHeight:      globalConfig.Config.ThumbMiniHeight,
									LeftRightButtonWidth: globalConfig.Config.LeftRightButtonWidth,
									ChapterLabelWidth:    globalConfig.Config.ChapterLabelWidth,
									ThumbnailWidth:       globalConfig.Config.ThumbnailWidth,
									ThumbnailHeight:      globalConfig.Config.ThumbnailHeight,
									ThumbTextHeight:      globalConfig.Config.ThumbTextHeight,
									NbWorkers:            globalConfig.Config.NbWorkers,
								},
								History: settings.History{
									Titles: globalConfig.History.Titles,
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

	title := canvas.NewText(si.MangaFound.Name, theme.TextColor())
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
	return fyne.NewSize(globalConfig.Config.ThumbnailWidth*globalConfig.Config.NbColumns, globalConfig.Config.ThumbMiniHeight+theme.Padding()*2)
}

func (s *SearchItemRender) Layout(size fyne.Size) {
	p := theme.Padding()

	s.lineUp.Resize(fyne.NewSize(globalConfig.Config.ThumbnailWidth*globalConfig.Config.NbColumns, 1))
	s.lineUp.Move(fyne.NewPos(0, 1))

	s.lineDown.Resize(fyne.NewSize(globalConfig.Config.ThumbnailWidth*globalConfig.Config.NbColumns, 1))
	s.lineDown.Move(fyne.NewPos(0, globalConfig.Config.ThumbMiniHeight+p*2))

	dx := p
	dy := p

	s.thumbnail.Resize(fyne.NewSize(globalConfig.Config.ThumbMiniWidth, globalConfig.Config.ThumbMiniHeight))
	s.thumbnail.Move(fyne.NewPos(dx, dy))
	dx = dx + p + globalConfig.Config.ThumbMiniWidth

	s.title.Resize(fyne.NewSize(globalConfig.Config.ThumbnailWidth*globalConfig.Config.NbColumns-p-dx, globalConfig.Config.ThumbTextHeight))
	s.title.Move(fyne.NewPos(dx, dy))
	dy = dy + globalConfig.Config.ThumbTextHeight + p

	s.description.Resize(fyne.NewSize(globalConfig.Config.ThumbnailWidth*globalConfig.Config.NbColumns-p-dx, globalConfig.Config.ThumbMiniHeight))
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
	for _, m := range globalConfig.History.Titles {
		if m.Title == manga.Title {
			return true
		}
	}
	return false
}
