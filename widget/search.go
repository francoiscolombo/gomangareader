package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/settings"
)

func showSearchResults(app fyne.App, win fyne.Window, search string, results []settings.Manga) {
	w := app.NewWindow(fmt.Sprintf("%s - results", search))
	w.SetContent(widgetSearchResults(app, win, w, results))
	w.Resize(fyne.NewSize(thWidth*3, thHeight*2))
	w.Show()
}

func widgetSearchResults(app fyne.App, win fyne.Window, w fyne.Window, result []settings.Manga) *fyne.Container {
	resultPanel := fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	for _, manga := range result {
		desc := widget.NewLabel(manga.Description)
		desc.Wrapping = fyne.TextWrapWord
		description := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5, thHeight)), widget.NewScrollContainer(desc))
		panelDetail := fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
			fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
				widget.NewLabelWithStyle("Alternate name:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle(manga.AlternateName, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
			),
			fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
				widget.NewLabelWithStyle("Year of release:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle(manga.YearOfRelease, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
			),
			fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
				widget.NewLabelWithStyle("Author & Artist:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle(fmt.Sprintf("%s & %s", manga.Author, manga.Artist), fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
			),
			fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
				widget.NewLabelWithStyle("Status:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabelWithStyle(manga.Status, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
			),
			description,
			widget.NewButtonWithIcon("Add this to my collection!", theme.ContentAddIcon(), func() {
				cnf := dialog.NewConfirm("Confirmation", fmt.Sprintf("Are you sure you want to\nadd \"%s\" to your collection?", manga.Name), func(b bool) {
					if !b {
						return
					}
					settings.UpdateHistory(*globalConfig, manga.Title, 1)
					win.SetTitle("GoMangaReader - Update in progress...")
					win.SetContent(fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(400, 20)),
						widget.NewLabelWithStyle("Please wait, refreshing your library...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
						widget.NewProgressBarInfinite()),
					)
					updateLibrary(app, win)
					dialog.ShowInformation("Information", fmt.Sprintf("The new serie \"%s\"\nis now part of your collection,\ncongratulations.", manga.Name), w)
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "GoMangaReader",
						Content: fmt.Sprintf("The new serie \"%s\" is now part of your library.\nCongratulations!", manga.Name),
					})
				}, w)
				cnf.SetDismissText("No, I changed my mind.")
				cnf.SetConfirmText("Of course I want!")
				cnf.Show()
			}),
		)
		resultPanel.AddObject(widget.NewAccordionContainer(widget.NewAccordionItem(manga.Name, panelDetail)))
	}
	return resultPanel
}
