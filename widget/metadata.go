package widget

import (
	"archive/zip"
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/dialog"
	"github.com/francoiscolombo/gomangareader/settings"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

/*
UpdateMetaData will retrieve all metadata directly from the provider
it will fill all the settings metadata but also download the cover of the serie, and generate a thumbnail
for every cbz existing (it will be the first page of the cbz)
all the additional graphics (serie cover, thumbnails) are generated on the following directory:
<manga path>/.meta/<manga title>-xxxx
the serie cover will be named "cover.jpg" and the thumbnail 000 - nb of chapter.jpg
*/
func UpdateMetaData(win fyne.Window, config settings.Settings) {
	// create metadata directory if it does not exists yet
	metadataPath := filepath.FromSlash(fmt.Sprintf("%s/.metadata", config.Config.LibraryPath))
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		err := os.MkdirAll(metadataPath, os.ModePerm)
		if err != nil {
			err := errors.New(fmt.Sprintf("Error while trying to create directory %s: %s", metadataPath, err))
			dialog.ShowError(err, win)
			fyne.CurrentApp().Quit()
		}
	}
	var mangaUpdatedList []settings.Manga
	for _, manga := range config.History.Titles {
		var provider settings.MangaProvider
		if manga.Provider == "mangareader.cc" {
			provider = settings.MangaReader{}
		}
		if provider != nil {
			newManga := provider.FindDetails(config.Config.LibraryPath, manga.Title, manga.LastChapter)
			provider.BuildChaptersList(&newManga)
			mangaUpdatedList = append(mangaUpdatedList, newManga)
			// download cover picture (if needed)
			downloadCover(win, newManga)
			// and generate thumbnails (if needed)
			extractFirstPages(win, config.Config.LibraryPath, newManga)
			//fmt.Println(" completed.")
		}
	}
	// okay we have updated the metadata, now we can save the config
	newSettings := settings.Settings{
		Config: settings.Config{
			LibraryPath: config.Config.LibraryPath,
		},
		History: settings.History{
			Titles: mangaUpdatedList,
		},
	}
	settings.WriteSettings(newSettings)
	//fmt.Println("> Settings updated.")
}

/*
DownloadCover simply download a cover for a manga title
*/
func downloadCover(win fyne.Window, manga settings.Manga) {
	// download only if does not exists
	if _, err := os.Stat(manga.CoverPath); os.IsNotExist(err) {
		//fmt.Printf("- %s does not exists yet, we have to download it....", manga.CoverPath)
		req, _ := http.NewRequest("GET", manga.CoverUrl, nil)
		req.Header.Add("cache-control", "no-cache")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			dialog.ShowError(err, win)
			fyne.CurrentApp().Quit()
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Something went wrong when I tried to clse the socket, the error is %s", err)
			}
		}(res.Body)
		if res.StatusCode != 200 {
			err := errors.New(fmt.Sprintf("status code error while trying to extract images from %s: %d %s\nclick OK to continue...", manga.CoverUrl, res.StatusCode, res.Status))
			dialog.ShowError(err, win)
		} else {
			//open a file for writing
			file, err := os.Create(manga.CoverPath)
			if err != nil {
				dialog.ShowError(err, win)
				fyne.CurrentApp().Quit()
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Printf("Something went wrong when I tried to clse the socket, the error is %s", err)
				}
			}(file)
			// Use io.Copy to just dump the response body to the file. This supports huge files
			_, err = io.Copy(file, res.Body)
			if err != nil {
				dialog.ShowError(err, win)
				fyne.CurrentApp().Quit()
			}
			//fmt.Println("- new cover downloaded")
		}
	}
}

/*
ExtractFirstPage allows to extract the first page of a cbz archive to generate a thumbnail
*/
func extractFirstPages(win fyne.Window, globalPath string, manga settings.Manga) {
	for i := 0; i < len(manga.Chapters); i++ {
		cbzArchive := filepath.FromSlash(fmt.Sprintf("%s/%s/%s-%03.1f.cbz", globalPath, manga.Title, manga.Title, manga.Chapters[i]))
		cbzThumbnail := filepath.FromSlash(fmt.Sprintf("%s/.metadata/%s-%03.1f.jpg", globalPath, manga.Title, manga.Chapters[i]))
		if _, err := os.Stat(cbzThumbnail); os.IsNotExist(err) {
			r, err := zip.OpenReader(cbzArchive)
			if err == nil {
				f := r.File[0]
				outFile, err := os.OpenFile(cbzThumbnail, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					msg := errors.New(fmt.Sprintf("Error while opening thumbnail %s: %s", cbzThumbnail, err))
					dialog.ShowError(msg, win)
					fyne.CurrentApp().Quit()
				}
				rc, err := f.Open()
				if err != nil {
					msg := errors.New(fmt.Sprintf("Error while opening first page of %s: %s", cbzThumbnail, err))
					dialog.ShowError(msg, win)
					fyne.CurrentApp().Quit()
				}
				_, err = io.Copy(outFile, rc)
				_ = outFile.Close()
				_ = rc.Close()
				if err != nil {
					msg := errors.New(fmt.Sprintf("Error while saving %s: %s", cbzThumbnail, err))
					dialog.ShowError(msg, win)
					fyne.CurrentApp().Quit()
				}
				err = r.Close()
				if err != nil {
					msg := errors.New(fmt.Sprintf("Error while closing archive %s: %s", cbzArchive, err))
					dialog.ShowError(msg, win)
					fyne.CurrentApp().Quit()
				}
			}
		}
	}
}
