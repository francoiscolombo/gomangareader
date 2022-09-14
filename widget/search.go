package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"sort"
)

type Search struct {
	widget.BaseWidget
	Provider         string
	Search           string
	Results          []settings.Manga
	SearchInProgress bool
}

func NewSearch(provider string) *Search {
	ns := &Search{
		BaseWidget:       widget.BaseWidget{},
		Provider:         provider,
		Search:           "",
		Results:          []settings.Manga{},
		SearchInProgress: false,
	}
	ns.ExtendBaseWidget(ns)
	return ns
}

// MinSize returns the size that this widget should not shrink below
func (s *Search) MinSize() fyne.Size {
	s.ExtendBaseWidget(s)
	return s.BaseWidget.MinSize()
}

func (s *Search) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)

	bg := canvas.NewRectangle(theme.ButtonColor())

	searchEntry := widget.NewEntry()

	searchForm := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "What title do you search?", Widget: searchEntry},
		},
		OnSubmit: func() {
			s.SearchInProgress = true
			s.Search = searchEntry.Text
			s.Refresh()
		},
		SubmitText: "Let's search for these titles!",
	}

	lblSearch := widget.NewLabel(fmt.Sprintf("Found %d results for search on '%s' with provider %s", len(s.Results), s.Search, s.Provider))
	lblSearch.Wrapping = fyne.TextWrapWord

	var results []*SearchItem

	sr := &SearchRenderer{
		bg:     bg,
		entry:  searchEntry,
		form:   searchForm,
		label:  lblSearch,
		items:  results,
		layout: nil,
		search: s,
	}

	return sr
}

type SearchRenderer struct {
	bg          *canvas.Rectangle
	entry       *widget.Entry
	form        *widget.Form
	label       *widget.Label
	items       []*SearchItem
	addSelected *widget.Button
	layout      fyne.Layout
	search      *Search
}

func (s *SearchRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (s *SearchRenderer) Destroy() {
	s.bg = nil
	s.entry = nil
	s.form = nil
	s.label = nil
	s.items = nil
	s.layout = nil
	s.search = nil
}

func (s *SearchRenderer) MinSize() fyne.Size {
	height := config.Config.ThumbTextHeight + theme.Padding()*2
	for _, i := range s.items {
		height = height + i.Size().Height
	}
	return fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns, height)
}

func (s *SearchRenderer) Layout(size fyne.Size) {
	p := theme.Padding()
	dx := p
	dy := p

	s.form.Resize(fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns, config.Config.ThumbTextHeight*3))
	s.form.Move(fyne.NewPos(dx, dy))
	dy = dy + config.Config.ThumbTextHeight*4 + p

	s.label.Resize(fyne.NewSize(config.Config.ThumbnailWidth*config.Config.NbColumns-p*2, config.Config.ThumbTextHeight))
	s.label.Move(fyne.NewPos(dx, dy))
	dy = dy + p + config.Config.ThumbTextHeight

	for _, i := range s.items {
		i.Resize(i.MinSize())
		i.Move(fyne.NewPos(dx, dy))
		dy = dy + p + config.Config.ThumbMiniHeight
	}
}

func (s *SearchRenderer) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	objects = append(objects, s.bg)
	objects = append(objects, s.form)
	objects = append(objects, s.label)
	for _, i := range s.items {
		objects = append(objects, i)
	}
	return objects
}

func (s *SearchRenderer) Refresh() {
	s.bg.Refresh()
	if s.search.SearchInProgress == true {
		s.search.SearchInProgress = false
		s.label = widget.NewLabel(fmt.Sprintf("Searching for '%s' with provider %s", s.search.Search, s.search.Provider))
		s.label.Wrapping = fyne.TextWrapWord
		s.label.Refresh()
		p := settings.MangaReader{}
		if s.search.Search != "" {
			s.search.Results = p.SearchManga(config.Config.LibraryPath, s.search.Search)
			sort.Slice(s.search.Results, func(i, j int) bool {
				return s.search.Results[i].Title < s.search.Results[j].Title
			})
		}
		s.items = []*SearchItem{}
		for _, r := range s.search.Results {
			item := NewSearchItem(r)
			s.items = append(s.items, item)
		}
		s.label = widget.NewLabel(fmt.Sprintf("Found %d results for search on '%s' with provider %s", len(s.items), s.search.Search, s.search.Provider))
		s.label.Wrapping = fyne.TextWrapWord
	}
	s.label.Refresh()
	for _, i := range s.items {
		i.Refresh()
	}
}
