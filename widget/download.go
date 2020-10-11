package widget

import (
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"github.com/francoiscolombo/gomangareader/settings"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func showDownloadChapters(app fyne.App, win fyne.Window, manga settings.Manga) {
	prog := dialog.NewProgress(fmt.Sprintf("Downloading new chapters for %s", manga.Name), fmt.Sprintf("Chapter %03d: download in progress....", manga.LastChapter), win)
	go func() {
		var provider settings.MangaProvider
		if globalConfig.Config.Provider == "mangareader.net" {
			provider = settings.MangaReader{}
		}
		imageLinks := provider.GetPagesUrls(manga)
		// okay now we have all images links, so download all the images... if we have any
		if len(imageLinks) > 0 {
			tempDirectory, err := ioutil.TempDir("", manga.Title)
			if err != nil {
				dialog.ShowError(err, win)
				fyne.CurrentApp().Quit()
			}
			for pageNumber, imgLink := range imageLinks {
				//log.Printf("Download page %d from url %s in temp directory %s\n",pageNumber,imgLink,tempDirectory)
				prog.SetValue(float64(pageNumber) / float64(len(imageLinks)))
				err := downloadImage(tempDirectory, pageNumber, imgLink)
				if err != nil {
					msg := errors.New(fmt.Sprintf("Issue when downloading page %d from url:\n%s\nthe error is: %s\nclick OK to continue...", pageNumber, imgLink, err))
					dialog.ShowError(msg, win)
				}
			}
			// and now create the new cbz from that temporary directory
			err = createCBZ(manga.Path, tempDirectory, manga.Title, manga.LastChapter)
			if err != nil {
				log.Panicln(err)
				dialog.ShowError(err, win)
				fyne.CurrentApp().Quit()
			}
		}
		prog.SetValue(1)
		// update history
		manga.LastChapter = manga.LastChapter + 1
		*globalConfig = settings.UpdateHistory(*globalConfig, manga.Title, manga.LastChapter)
		prog.Hide()
		if manga.LastChapter <= provider.CheckLastChapter(manga) {
			showDownloadChapters(app, win, manga)
		} else {
			updateLibrary(app, win)
		}
	}()
	prog.Show()
}

func downloadImage(path string, page int, url string) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Add("cache-control", "no-cache")
	// Create a new HTTP client and execute the request
	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("status code error while trying to extract images from %s: %d %s", url, res.StatusCode, res.Status))
	} else {
		defer res.Body.Close()
		//open a file for writing
		file, err := os.Create(filepath.FromSlash(fmt.Sprintf("%s/page_%03d.jpg", path, page)))
		if err != nil {
			return err
		}
		defer file.Close()
		// Use io.Copy to just dump the response body to the file. This supports huge files
		_, err = io.Copy(file, res.Body)
		if err != nil {
			return err
		}
	}
	return nil
}
