package widget

import (
	"archive/zip"
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/francoiscolombo/gomangareader/settings"
	"io"
	"log"
	"net/http"
	"os"
	"path"
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
			err := downloadCover(newManga)
			if err != nil {
				dialog.ShowError(err, win)
			}
			// and generate thumbnails (if needed)
			err = extractFirstPages(config.Config.LibraryPath, newManga)
			if err != nil {
				dialog.ShowError(err, win)
			}
			//fmt.Println(" completed.")
		}
	}
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
			Titles: mangaUpdatedList,
		},
	}
	settings.WriteSettings(newSettings)
	//fmt.Println("> Settings updated.")
}

/*
DownloadCover simply download a cover for a manga title
*/
func downloadCover(manga settings.Manga) error {
	// download only if does not exists
	if _, err := os.Stat(manga.CoverPath); os.IsNotExist(err) {
		// first ensure that the path exists...
		err := os.MkdirAll(path.Dir(manga.CoverPath), 0750)
		if err != nil {
			return err
		}
		//fmt.Printf("- %s does not exists yet, we have to download it....", manga.CoverPath)
		req, _ := http.NewRequest("GET", manga.CoverUrl, nil)
		req.Header.Add("cache-control", "no-cache")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Something went wrong when I tried to clse the socket, the error is %s", err)
			}
		}(res.Body)
		if res.StatusCode != 200 {
			err := errors.New(fmt.Sprintf("status code error while trying to extract images from %s: %d %s\nclick OK to continue...", manga.CoverUrl, res.StatusCode, res.Status))
			return err
		} else {
			//open a file for writing
			file, err := os.Create(manga.CoverPath)
			if err != nil {
				return err
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					log.Printf("Something went wrong when I tried to close the socket, the error is %s", err)
				}
			}(file)
			// Use io.Copy to just dump the response body to the file. This supports huge files
			_, err = io.Copy(file, res.Body)
			if err != nil {
				return err
			}
			//log.Println("- new cover downloaded")
		}
	}
	return nil
}

/*
ExtractFirstPage allows to extract the first page of a cbz archive to generate a thumbnail
*/
func extractFirstPages(globalPath string, manga settings.Manga) error {
	for i := 0; i < len(manga.Chapters); i++ {
		cbzArchive := filepath.FromSlash(fmt.Sprintf("%s/%s/%s-%03.1f.cbz", globalPath, manga.Title, manga.Title, manga.Chapters[i]))
		cbzThumbnail := filepath.FromSlash(fmt.Sprintf("%s/.metadata/%s-%03.1f.jpg", globalPath, manga.Title, manga.Chapters[i]))
		if _, err := os.Stat(cbzThumbnail); os.IsNotExist(err) {
			r, err := zip.OpenReader(cbzArchive)
			if err == nil {
				if len(r.File) == 0 {
					log.Printf("Error while trying to get first pages for manga %s, chapter %03.1f: no pages to extract. Process abandonned, pass to next one.", manga.Name, manga.Chapters[i])
					break
				}
				f := r.File[0]
				outFile, err := os.OpenFile(cbzThumbnail, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					return errors.New(fmt.Sprintf("Error while opening thumbnail %s: %s", cbzThumbnail, err))
				}
				rc, err := f.Open()
				if err != nil {
					return errors.New(fmt.Sprintf("Error while opening first page of %s: %s", cbzThumbnail, err))
				}
				_, err = io.Copy(outFile, rc)
				_ = outFile.Close()
				_ = rc.Close()
				if err != nil {
					return errors.New(fmt.Sprintf("Error while saving %s: %s", cbzThumbnail, err))
				}
				err = r.Close()
				if err != nil {
					return errors.New(fmt.Sprintf("Error while closing archive %s: %s", cbzArchive, err))
				}
			}
		}
	}
	return nil
}
