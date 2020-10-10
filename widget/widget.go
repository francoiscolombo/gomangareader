package widget

import (
	"archive/zip"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/PuerkitoBio/goquery"
	"github.com/francoiscolombo/gomangareader/archive"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	thWidth   = 144
	thHeight  = 192
	chWidth   = 48
	chHeight  = 64
	pgWidth   = 600
	pgHeight  = 585
	btnHeight = 30
)

var globalConfig *settings.Settings

/*
Library allow to display the mangas in a GUI.
*/
func ShowLibrary() {

	libraryViewer := app.NewWithID("gomangareader")
	libraryViewer.SetIcon(theme.FyneLogo())

	windows := libraryViewer.NewWindow("GoMangaReader - Update in progress...")
	windows.SetContent(fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(400, 20)),
		widget.NewLabelWithStyle("Please wait, loading your library...", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true}),
		widget.NewProgressBarInfinite()),
	)
	windows.Show()

	go func() {
		updateLibrary(libraryViewer, windows)
	}()

	libraryViewer.Run()
}

func updateLibrary(app fyne.App, win fyne.Window) {
	win.SetTitle("GoMangaReader - Update in progress...")
	if settings.IsSettingsExisting() == false {
		settings.WriteDefaultSettings()
	}
	cfg := settings.ReadSettings()
	globalConfig = &cfg
	fmt.Println("- Settings loaded.")
	fmt.Printf("  > Library path is %s\n  > Default provider is %s\n\n", globalConfig.Config.LibraryPath, globalConfig.Config.Provider)
	settings.UpdateMetaData(*globalConfig)
	cfg = settings.ReadSettings()
	globalConfig = &cfg
	content := fyne.NewContainerWithLayout(layout.NewGridLayout(5))
	for _, manga := range globalConfig.History.Titles {
		ws := widgetSerie(app, win, manga)
		if ws != nil {
			ws.Resize(fyne.NewSize(thWidth, thHeight))
			content.AddObject(ws)
		}
	}
	win.SetContent(fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		widgetUpdateCollections(app, win),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5+40, (thHeight+btnHeight*2)*2)),
			widget.NewScrollContainer(content))))
	win.SetTitle("GoMangaReader")
}

func contains(array []string, value string) bool {
	for i := 0; i < len(array); i++ {
		if array[i] == value {
			return true
		}
	}
	return false
}

func showDownloadChapters(app fyne.App, win fyne.Window, manga settings.Manga) {
	prog := dialog.NewProgress(fmt.Sprintf("Downloading new chapters for %s", manga.Name), fmt.Sprintf("Chapter %03d: download in progress....", manga.LastChapter), win)
	go func() {
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangareader.net/%s/%d", manga.Title, manga.LastChapter), nil)
		req.Header.Add("cache-control", "no-cache")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			// handle error
			log.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Fatalf("status code error while trying to get details for title %s: %d %s", manga.Title, res.StatusCode, res.Status)
		}
		body, _ := ioutil.ReadAll(res.Body)
		// we have to analyze the body since the images are located in a script
		content := strings.Split(string(body), "\"")
		var imageLinks []string
		for _, val := range content {
			if strings.Contains(val, manga.Title) && strings.HasSuffix(val, ".jpg") {
				link := strings.ReplaceAll(val, "\\/", "/")
				if !strings.HasPrefix(val, "http") {
					link = fmt.Sprintf("https:%s", link)
				}
				if !contains(imageLinks, link) {
					imageLinks = append(imageLinks, link)
				}
			}
		}
		// okay now we have all images links, so download all the images... if we have any
		if len(imageLinks) > 0 {
			tempDirectory, err := ioutil.TempDir("", manga.Title)
			if err != nil {
				log.Fatal(err)
			}
			for pageNumber, imgLink := range imageLinks {
				//log.Printf("Download page %d from url %s in temp directory %s\n",pageNumber,imgLink,tempDirectory)
				prog.SetValue(float64(pageNumber) / float64(len(imageLinks)))
				downloadImage(tempDirectory, pageNumber, imgLink)
			}
			// and now create the new cbz from that temporary directory
			createCBZ(manga.Path, tempDirectory, manga.Title, manga.LastChapter)
		}
		prog.SetValue(1)
		// update history
		manga.LastChapter = manga.LastChapter + 1
		*globalConfig = settings.UpdateHistory(*globalConfig, manga.Title, manga.LastChapter)
		prog.Hide()
		if manga.LastChapter <= getLastChapter(manga) {
			showDownloadChapters(app, win, manga)
		} else {
			updateLibrary(app, win)
		}
	}()
	prog.Show()
}

func searchManga(search string) (result []settings.Manga) {
	req, _ := http.NewRequest("GET", "https://www.mangareader.net/alphabetical", nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to find title %s: %d %s", search, res.StatusCode, res.Status)
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	doc.Find("a").Each(func(i int, link *goquery.Selection) {
		l, _ := link.Attr("href")
		v := link.Text()
		if strings.HasPrefix(l, "/") {
			l = l[1:]
		}
		if strings.Contains(strings.ToLower(v), strings.ToLower(search)) {
			found := settings.UpdateDetails(*globalConfig, settings.Manga{Title: l})
			result = append(result, found)
		}
		return
	})
	return
}

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

func widgetUpdateCollections(app fyne.App, win fyne.Window) *fyne.Container {
	return fyne.NewContainerWithLayout(layout.NewGridLayout(2),
		widget.NewButtonWithIcon("Search series to add to the collection", theme.SearchIcon(), func() {
			newSerie := widget.NewEntry()
			content := widget.NewForm(
				widget.NewFormItem("Search series to add:", newSerie),
			)
			dialog.ShowCustomConfirm("What do you want to find?", "Search", "Cancel", content, func(b bool) {
				if !b {
					return
				}
				// and here we have to add it.
				searchResults := searchManga(newSerie.Text)
				showSearchResults(app, win, newSerie.Text, searchResults)
			}, win)
		}),
		widget.NewButtonWithIcon("Update collection", theme.DocumentIcon(), func() {
			cnf := dialog.NewConfirm("Confirmation", "Are you sure you want to\nupdate your collection?", func(b bool) {
				if !b {
					return
				}
				updateLibrary(app, win)
				canvas.Refresh(win.Canvas().Content())
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "GoMangaReader",
					Content: "Your library is now updated.",
				})
			}, win)
			cnf.SetDismissText("Nah")
			cnf.SetConfirmText("Oh Yes!")
			cnf.Show()
		}),
	)
}

func widgetSerie(app fyne.App, win fyne.Window, manga settings.Manga) *fyne.Container {
	title := manga.Name
	if len(title) > 13 {
		title = title[0:13] + "..."
	}
	colorTitle := color.NRGBA{0x80, 0xff, 0, 0xff}
	if checkNewChapters(manga) {
		colorTitle = color.NRGBA{0xff, 0x80, 0, 0xff}
	}
	return fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, thHeight)), canvas.NewImageFromFile(manga.CoverPath)),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, 20)), canvas.NewText(fmt.Sprintf("%s (%d)", title, manga.LastChapter-1), colorTitle)),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, btnHeight)), widget.NewButton("Show", func() {
			showSerieDetail(app, manga)
		})))
}

func showSerieDetail(app fyne.App, manga settings.Manga) {
	w := app.NewWindow(fmt.Sprintf("%s - details", manga.Name))
	w.SetContent(widgetDetailSerie(app, w, manga))
	w.Resize(fyne.NewSize(thWidth*3, thHeight*2))
	w.Show()
}

func chaptersPanel(app fyne.App, win fyne.Window, manga settings.Manga, available bool) *fyne.Container {
	var chaps []string
	currentChapter := app.Preferences().Int(manga.Title)
	if currentChapter <= 0 {
		currentChapter = 1
	}
	nbChapters := manga.LastChapter - 1
	for i := 1; i <= nbChapters; i++ {
		chaps = append(chaps, fmt.Sprintf("Chapter %d / %d", i, nbChapters))
	}
	thumbnailPath := filepath.Dir(manga.CoverPath)

	thumbnailView := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	thumbnailView.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.jpg", thumbnailPath, manga.Title, currentChapter))
	canvas.Refresh(thumbnailView)

	selChapter := widget.NewSelect(chaps, func(s string) {
		f := strings.Fields(s)
		currentChapter, _ = strconv.Atoi(f[1])
		thumbnailView.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.jpg", thumbnailPath, manga.Title, currentChapter))
		canvas.Refresh(thumbnailView)
	})
	selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d", currentChapter, nbChapters))
	prev := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		currentChapter--
		if currentChapter < 1 {
			currentChapter = 1
		}
		selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d", currentChapter, nbChapters))
	})
	next := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		currentChapter++
		if currentChapter > nbChapters {
			currentChapter = nbChapters
		}
		selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d", currentChapter, nbChapters))
	})
	read := widget.NewButtonWithIcon("Read this chapter...", theme.DocumentIcon(), func() {
		readChapter(app, manga, currentChapter)
	})
	if available {
		read.Text = "Read this..."
		download := widget.NewButtonWithIcon("Download chapters...", theme.DocumentSaveIcon(), func() {
			showDownloadChapters(app, win, manga)
			/*
				w := app.NewWindow(fmt.Sprintf("%s - download new chapters",manga.Name))
				label := widget.NewLabel(fmt.Sprintf("Download new chapters for %s...",manga.Name))
				chapterInProgress := canvas.NewText("downloading chapter ##", color.NRGBA{0xff, 0xc0, 0x80, 0xff})
				chapterInProgress.Alignment = fyne.TextAlignCenter
				pBar := widget.NewProgressBar()
				pBar.Min = float64(manga.LastChapter)
				pBar.Max = float64(getLastChapter(manga))
				pBar.SetValue(0.0)
				content := fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
					fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*3,30)),label),
					fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*3,30)),chapterInProgress),
					fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*3,30)),pBar))
				w.SetContent(content)
				w.Resize(fyne.NewSize(thWidth*3,100))
				go func() {
					c := manga.LastChapter
					for {
						chapterInProgress.Text = fmt.Sprintf("downloading chapter %03d",c)
						canvas.Refresh(chapterInProgress)
						pBar.SetValue(float64(c))
						canvas.Refresh(pBar)
						downloadChapter(app, win, manga,c)
						c = c + 1
					}
					updateLibrary(app,win)
					w.Close()
				}()
				w.Show()
			*/

		})
		return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2, chHeight)), prev),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth, chHeight)), thumbnailView),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2, chHeight)), next),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*1.5, chHeight)), selChapter),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, chHeight)), read),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*1.5, chHeight)), download))
	}
	return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2, chHeight)), prev),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth, chHeight)), thumbnailView),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2, chHeight)), next),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2, chHeight)), selChapter),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2, chHeight)), read))
}

func widgetDetailSerie(app fyne.App, win fyne.Window, manga settings.Manga) *fyne.Container {
	coverImage := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2, thHeight*2)), canvas.NewImageFromFile(manga.CoverPath))
	updated := checkNewChapters(manga)
	nca := canvas.NewText("complete", color.NRGBA{0x80, 0x80, 0xff, 0xff})
	if updated {
		nca = canvas.NewText("new chapters available", color.NRGBA{0xff, 0x80, 0x80, 0xff})
	}
	availability := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2, 30)), nca)
	detailPanel := fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, 30)),
			widget.NewLabel("Name:"),
			widget.NewLabel("Alternate Name:"),
			widget.NewLabel("Year of release:"),
			widget.NewLabel("Status:"),
			widget.NewLabel("Nb of Chapters:"),
			widget.NewLabel("Reading direction:"),
			widget.NewLabel("Author:"),
			widget.NewLabel("Artist:"),
			availability),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth, 30)),
			widget.NewLabel(manga.Name),
			widget.NewLabel(manga.AlternateName),
			widget.NewLabel(manga.YearOfRelease),
			widget.NewLabel(manga.Status),
			widget.NewLabel(fmt.Sprintf("%d", manga.LastChapter-1)),
			widget.NewLabel(manga.ReadingDirection),
			widget.NewLabel(manga.Author),
			widget.NewLabel(manga.Artist),
			widget.NewLabel("")))
	desc := widget.NewLabel(manga.Description)
	desc.Wrapping = fyne.TextWrapWord
	description := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5, thHeight)), widget.NewScrollContainer(desc))
	browseChapters := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5, chHeight)), chaptersPanel(app, win, manga, updated))
	return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewVBoxLayout(), fyne.NewContainerWithLayout(layout.NewHBoxLayout(), coverImage, detailPanel), browseChapters, description))
}

func readChapter(app fyne.App, manga settings.Manga, chapter int) {
	w := app.NewWindow(fmt.Sprintf("Read %s - chapter %d", manga.Name, chapter))
	w.SetContent(widget.NewScrollContainer(widgetReader(app, manga, chapter)))
	w.Resize(fyne.NewSize(pgWidth, pgHeight))
	w.Show()
	w.SetOnClosed(func() {
		metadataPath := filepath.Dir(manga.CoverPath)
		tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer", metadataPath, manga.Title))
		err := os.RemoveAll(tmpDir)
		if err != nil {
			fmt.Printf("Error when trying to remove temporary view folder: %s", err)
		}
	})
}

func widgetReader(app fyne.App, manga settings.Manga, chapter int) *fyne.Container {
	app.Preferences().SetInt(manga.Title, chapter)
	metadataPath := filepath.Dir(manga.CoverPath)
	tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer", metadataPath, manga.Title))
	cbzPath := filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.cbz", manga.Path, manga.Title, chapter))
	pages, err := archive.Unzip(cbzPath, tmpDir)
	if err != nil {
		fmt.Printf("Error when trying to unzip %s to temporary view folder %s: %s", cbzPath, tmpDir, err)
	}
	//pageNumber := app.Preferences().Int(fmt.Sprintf("%s/page/",manga.Title))
	//if pageNumber <= 0 {
	//	pageNumber = 1
	//}
	pageNumber := 1
	nbPages := len(pages)

	displayPage := widget.NewLabel(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
	pageProgress := widget.NewProgressBar()
	pageProgress.SetValue(float64(pageNumber) / float64(nbPages))

	pageView := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	pageView.File = pages[pageNumber-1]
	canvas.Refresh(pageView)

	navBar := fyne.NewContainerWithLayout(
		layout.NewGridLayout(2),
		widget.NewHBox(
			widget.NewButtonWithIcon("[Prev]", theme.MediaFastRewindIcon(), func() {
				pageNumber--
				if pageNumber < 1 {
					pageNumber = 1
				}
				displayPage.SetText(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
				pageProgress.SetValue(float64(pageNumber) / float64(nbPages))
				pageView.File = pages[pageNumber-1]
				canvas.Refresh(pageView)
				app.Preferences().SetInt(fmt.Sprintf("%s/page/", manga.Title), pageNumber)
			}),
			widget.NewButtonWithIcon("[Next]", theme.MediaFastForwardIcon(), func() {
				pageNumber++
				if pageNumber > nbPages {
					pageNumber = nbPages
				}
				displayPage.SetText(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
				pageProgress.SetValue(float64(pageNumber) / float64(nbPages))
				pageView.File = pages[pageNumber-1]
				canvas.Refresh(pageView)
				app.Preferences().SetInt(fmt.Sprintf("%s/page/", manga.Title), pageNumber)
			}),
			layout.NewSpacer(),
			displayPage,
		),
		pageProgress,
	)

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, navBar, nil, nil), navBar, pageView)
}

func getLastChapter(manga settings.Manga) int {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangareader.net/%s", manga.Title), nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to get details for title %s: %d %s", manga.Title, res.StatusCode, res.Status)
	}
	body, _ := ioutil.ReadAll(res.Body)
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		log.Fatal(err)
	}
	// check available chapters
	var lastChapter int
	doc.Find("a").Each(func(i int, link *goquery.Selection) {
		v, _ := link.Attr("href")
		if strings.HasPrefix(v, "/"+manga.Title+"/") {
			s := strings.Split(v, "/")
			lastChapter, _ = strconv.Atoi(s[2])
		}
		return
	})
	return lastChapter
}

func checkNewChapters(manga settings.Manga) bool {
	return getLastChapter(manga) >= manga.LastChapter
}

func downloadImage(path string, page int, url string) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("cache-control", "no-cache")
	// Create a new HTTP client and execute the request
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	if res.StatusCode != 200 {
		log.Printf("status code error while trying to extract images from %s: %d %s", url, res.StatusCode, res.Status)
	} else {
		defer res.Body.Close()
		//open a file for writing
		file, err := os.Create(filepath.FromSlash(fmt.Sprintf("%s/page_%03d.jpg", path, page)))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		// Use io.Copy to just dump the response body to the file. This supports huge files
		_, err = io.Copy(file, res.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createCBZ(outputPath, pagesPath, title string, chapter int) error {
	// List of Files to Zip
	fmt.Printf("\ncreate %s ... ", fmt.Sprintf("%s-%03d.cbz", title, chapter))
	var files []string
	outputCBZ := filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.cbz", outputPath, title, chapter))
	err := filepath.Walk(pagesPath, func(path string, info os.FileInfo, err error) error {
		src, err := os.Stat(path)
		if err != nil {
			// still does not exists? then something wrong, exit in panic mode.
			panic(err)
		}
		if !src.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	// create archive
	newZipFile, err := os.Create(outputCBZ)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		} else {
			os.Remove(file)
		}
	}

	// remove temporary folder
	os.Remove(pagesPath)
	fmt.Println("done")

	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	//header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
