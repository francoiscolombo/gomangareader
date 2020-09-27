package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/PuerkitoBio/goquery"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/francoiscolombo/gomangareader/settings"
	"github.com/francoiscolombo/gomangareader/util"
	"github.com/francoiscolombo/gomangareader/archive"
)

const (
	thWidth  = 144
	thHeight = 192
	chWidth  = 48
	chHeight = 64
	pgWidth  = 600
	pgHeight = 585
	btnHeight = 30
)

var globalConfig *settings.Settings

/*
Library allow to display the mangas in a GUI.
*/
func ShowLibrary(cfg *settings.Settings) {
	globalConfig = cfg
	libraryViewer := app.NewWithID("gomangareaderdl")
	w := libraryViewer.NewWindow("Your manga library")
	content := fyne.NewContainerWithLayout(layout.NewGridLayout(5))
	for _, manga := range cfg.History.Titles {
		b := widgetSerie(libraryViewer, manga)
		if b != nil {
			b.Resize(fyne.NewSize(thWidth,thHeight))
			content.AddObject(b)
		}
	}
	w.SetContent(fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		widget.NewButtonWithIcon("Update collections", theme.ContentRedoIcon(), func() {
			fmt.Println("not implemented yet")
		}),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5+40,(thHeight+btnHeight*2)*2)),
			widget.NewScrollContainer(content))))
	w.ShowAndRun()
}

func widgetSerie(app fyne.App, manga settings.Manga) *fyne.Container {
	title := manga.Name
	if len(title) > 13 {
		title = title[0:13] + "..."
	}
	return fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth,thHeight)),canvas.NewImageFromFile(manga.CoverPath)),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth,20)),canvas.NewText(fmt.Sprintf("%s (%d)",title,manga.LastChapter-1), color.NRGBA{0xff, 0x80, 0, 0xff})),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth,btnHeight)),widget.NewButton("Show", func() {
			showSerieDetail(app, manga)
		})))
}

func showSerieDetail(app fyne.App, manga settings.Manga) {
	w := app.NewWindow(fmt.Sprintf("%s - details",manga.Name))
	w.SetContent(widgetDetailSerie(app, manga))
	w.Resize(fyne.NewSize(thWidth*3,thHeight*2))
	w.Show()
}

func chaptersPanel(app fyne.App, manga settings.Manga, available bool) *fyne.Container {
	var chaps []string
	currentChapter := app.Preferences().Int(manga.Title)
	if currentChapter <= 0 {
		currentChapter = 1
	}
	nbChapters := manga.LastChapter - 1
	for i:=1; i<=nbChapters; i++ {
		chaps = append(chaps, fmt.Sprintf("Chapter %d / %d",i,nbChapters))
	}
	thumbnailPath := filepath.Dir(manga.CoverPath)

	thumbnailView := &canvas.Image{FillMode: canvas.ImageFillOriginal}
	thumbnailView.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.jpg",thumbnailPath,manga.Title,currentChapter))
	canvas.Refresh(thumbnailView)

	selChapter := widget.NewSelect(chaps, func(s string) {
		f := strings.Fields(s)
		currentChapter,_ = strconv.Atoi(f[1])
		thumbnailView.File = filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.jpg",thumbnailPath,manga.Title,currentChapter))
		canvas.Refresh(thumbnailView)
	})
	selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d",currentChapter,nbChapters))
	prev := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() {
		currentChapter--
		if currentChapter < 1 {
			currentChapter = 1
		}
		selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d",currentChapter,nbChapters))
	})
	next := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		currentChapter++
		if currentChapter > nbChapters {
			currentChapter = nbChapters
		}
		selChapter.SetSelected(fmt.Sprintf("Chapter %d / %d",currentChapter,nbChapters))
	})
	read := widget.NewButtonWithIcon("Read this chapter...", theme.DocumentIcon(), func() {
		readChapter(app,manga,currentChapter)
	})
	if available {
		read.Text = "Read this..."
		download := widget.NewButtonWithIcon("Download chapters...", theme.DocumentSaveIcon(), func() {
			w := app.NewWindow(fmt.Sprintf("%s - download new chapters",manga.Name))
			label := widget.NewLabel(fmt.Sprintf("Download new chapters for %s...",manga.Name))
			chapterInProgress := canvas.NewText("downloading chapter ##", color.NRGBA{0xff, 0xc0, 0x80, 0xff})
			chapterInProgress.Alignment = fyne.TextAlignCenter
			pBar := widget.NewProgressBar()
			pBar.Min = 0.0
			pBar.Max = float64(getLastChapter(manga) - manga.LastChapter)
			pBar.SetValue(0.0)
			content := fyne.NewContainerWithLayout(layout.NewVBoxLayout(),
				fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*3,30)),label),
				fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*3,30)),chapterInProgress),
				fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*3,30)),pBar))
			w.SetContent(content)
			w.Resize(fyne.NewSize(thWidth*3,100))
			go func() {
				c := manga.LastChapter
				for i:=pBar.Min; i<=pBar.Max; i++ {
					chapterInProgress.Text = fmt.Sprintf("downloading chapter %3.0f",i+float64(manga.LastChapter))
					canvas.Refresh(chapterInProgress)
					pBar.SetValue(i)
					canvas.Refresh(pBar)
					downloadChapter(manga,c)
					c = c + 1
				}
				UpdateMetaData(*globalConfig)
				w.Close()
			}()
			w.Show()
		})
		return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2,chHeight)),prev),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth,chHeight)),thumbnailView),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2,chHeight)),next),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*1.5,chHeight)),selChapter),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth,chHeight)),read),
			fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*1.5,chHeight)),download))
	}
	return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2,chHeight)),prev),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth,chHeight)),thumbnailView),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(chWidth/2,chHeight)),next),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2,chHeight)),selChapter),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2,chHeight)),read))
}

func widgetDetailSerie(app fyne.App, manga settings.Manga) *fyne.Container {
	coverImage := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2,thHeight*2)),canvas.NewImageFromFile(manga.CoverPath))
	updated := checkNewChapters(manga)
	nca := canvas.NewText("complete", color.NRGBA{0x80, 0x80, 0xff, 0xff})
	if updated {
		nca = canvas.NewText("new chapters available", color.NRGBA{0xff, 0x80, 0x80, 0xff})
	}
	availability := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*2,30)),nca)
	detailPanel := fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth,30)),
			widget.NewLabel("Name:"),
			widget.NewLabel("Alternate Name:"),
			widget.NewLabel("Year of release:"),
			widget.NewLabel("Status:"),
			widget.NewLabel("Nb of Chapters:"),
			widget.NewLabel("Reading direction:"),
			widget.NewLabel("Author:"),
			widget.NewLabel("Artist:"),
			availability),
		fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth,30)),
			widget.NewLabel(manga.Name),
			widget.NewLabel(manga.AlternateName),
			widget.NewLabel(manga.YearOfRelease),
			widget.NewLabel(manga.Status),
			widget.NewLabel(fmt.Sprintf("%d",manga.LastChapter-1)),
			widget.NewLabel(manga.ReadingDirection),
			widget.NewLabel(manga.Author),
			widget.NewLabel(manga.Artist),
			widget.NewLabel("")))
	desc := widget.NewLabel(manga.Description)
	desc.Wrapping = fyne.TextWrapWord
	description := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5,thHeight)),widget.NewScrollContainer(desc))
	browseChapters := fyne.NewContainerWithLayout(layout.NewGridWrapLayout(fyne.NewSize(thWidth*5,chHeight)),chaptersPanel(app,manga,updated))
	return fyne.NewContainerWithLayout(layout.NewHBoxLayout(),
		fyne.NewContainerWithLayout(layout.NewVBoxLayout(), fyne.NewContainerWithLayout(layout.NewHBoxLayout(),coverImage,detailPanel), browseChapters, description))
}

func readChapter(app fyne.App, manga settings.Manga, chapter int) {
	w := app.NewWindow(fmt.Sprintf("Read %s - chapter %d",manga.Name,chapter))
	w.SetContent(widget.NewScrollContainer(widgetReader(app, manga, chapter)))
	w.Resize(fyne.NewSize(pgWidth,pgHeight))
	w.Show()
	w.SetOnClosed(func() {
		metadataPath := filepath.Dir(manga.CoverPath)
		tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer",metadataPath,manga.Title))
		err := os.RemoveAll(tmpDir)
		if err != nil {
			fmt.Printf("Error when trying to remove temporary view folder: %s", err)
		}
	})
}

func widgetReader(app fyne.App, manga settings.Manga, chapter int) *fyne.Container {
	app.Preferences().SetInt(manga.Title,chapter)
	metadataPath := filepath.Dir(manga.CoverPath)
	tmpDir := filepath.FromSlash(fmt.Sprintf("%s/%s/viewer",metadataPath,manga.Title))
	cbzPath := filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.cbz",manga.Path,manga.Title,chapter))
	pages, err := Unzip(cbzPath, tmpDir)
	if err != nil {
		fmt.Printf("Error when trying to unzip %s to temporary view folder %s: %s", cbzPath, tmpDir, err)
	}
	pageNumber := app.Preferences().Int(fmt.Sprintf("%s/page/",manga.Title))
	if pageNumber <= 0 {
		pageNumber = 1
	}
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
			widget.NewButton("<<< [Prev]", func() {
				pageNumber--
				if pageNumber < 1 {
					pageNumber = 1
				}
				displayPage.SetText(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
				pageProgress.SetValue(float64(pageNumber) / float64(nbPages))
				pageView.File = pages[pageNumber-1]
				canvas.Refresh(pageView)
				app.Preferences().SetInt(fmt.Sprintf("%s/page/",manga.Title),pageNumber)
			}),
			widget.NewButton("[Next] >>>", func() {
				pageNumber++
				if pageNumber > nbPages {
					pageNumber = nbPages
				}
				displayPage.SetText(fmt.Sprintf("Page %d / %d", pageNumber, nbPages))
				pageProgress.SetValue(float64(pageNumber) / float64(nbPages))
				pageView.File = pages[pageNumber-1]
				canvas.Refresh(pageView)
				app.Preferences().SetInt(fmt.Sprintf("%s/page/",manga.Title),pageNumber)
			}),
			layout.NewSpacer(),
			displayPage,
		),
		pageProgress,
	)

	return fyne.NewContainerWithLayout(layout.NewBorderLayout(nil, navBar, nil, nil), navBar, pageView)
}

func getLastChapter(manga settings.Manga) int {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangareader.net/%s",manga.Title), nil)
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
		if strings.HasPrefix(v,"/"+manga.Title+"/") {
			s := strings.Split(v,"/")
			lastChapter,_ = strconv.Atoi(s[2])
		}
		return
	})
	return lastChapter
}

func checkNewChapters(manga settings.Manga) bool {
	return getLastChapter(manga) >= manga.LastChapter
}

func downloadChapter(manga settings.Manga, chapter int) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.mangareader.net/%s/%d",manga.Title,chapter), nil)
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
	content := strings.Split(string(body),"\"")
	var imageLinks []string
	for _, val := range content {
		if strings.Contains(val,manga.Title) && strings.HasSuffix(val,".jpg") {
			if strings.HasPrefix(val,"https:") || strings.HasPrefix(val,"http:") {
				imageLinks = append(imageLinks, strings.ReplaceAll(val,"\\/","/"))
			} else {
				imageLinks = append(imageLinks, fmt.Sprintf("https:%s",strings.ReplaceAll(val,"\\/","/")))
			}
		}
	}
	// okay now we have all images links, so download all the images
	tempDirectory, err := ioutil.TempDir("", manga.Title)
	if err != nil {
		log.Fatal(err)
	}
	for pageNumber, imgLink := range imageLinks {
		downloadImage(tempDirectory,pageNumber,imgLink)
	}
	// and now create the new cbz from that temporary directory
	createCBZ(manga.Path, tempDirectory, manga.Title, chapter)
}

func downloadImage(path string, page int, url string) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("cache-control", "no-cache")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		// handle error
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error while trying to extract images from %s: %d %s", url, res.StatusCode, res.Status)
	}
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

func createCBZ(outputPath, pagesPath, title string, chapter int) {
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
	if err := ZipFiles(outputCBZ, files); err != nil {
		panic(err)
	}
	for _, file := range files {
		os.Remove(file)
	}
	os.Remove(pagesPath)
	fmt.Println("done")
}
