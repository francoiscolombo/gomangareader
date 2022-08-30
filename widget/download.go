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
	"sync"
	"time"
)

type downloadWorkerJob struct {
	path string
	page int
	url  string
}

type downloadWorkerResult struct {
	error      error
	url        string
	statusCode int
}

func downloadWorker(jobs <-chan downloadWorkerJob, results chan<- downloadWorkerResult) {
	for workerJob := range jobs {
		//log.Printf("Start download page %d from url %s in temp directory %s\n",workerJob.page,workerJob.url,workerJob.path)
		status, err := downloadImage(workerJob.path, workerJob.page, workerJob.url)
		//log.Printf("Page %d downloaded with status code %d\n",workerJob.page,status)
		res := downloadWorkerResult{error: err, url: workerJob.url, statusCode: status}
		results <- res
	}
}

func downloadImage(path string, page int, url string) (int, error) {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url)
	if err != nil {
		return -1, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Something went wront while trying to close http client, error is %s", err)
		}
	}(resp.Body)

	out, err := os.Create(filepath.FromSlash(fmt.Sprintf("%s/page_%03d.jpg", path, page)))
	if err != nil {
		return -1, err
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			log.Printf("Something went wront while trying to close http client, error is %s", err)
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	return resp.StatusCode, err
}

func showDownloadChapters(app fyne.App, win fyne.Window, manga settings.Manga) {
	prog := dialog.NewProgress(fmt.Sprintf("Downloading new chapters for %s", manga.Name), fmt.Sprintf("Chapter %03.1f: download in progress....", manga.LastChapter), win)
	var provider settings.MangaProvider
	if manga.Provider == "mangareader.cc" {
		provider = settings.MangaReader{}
	}
	imageLinks := provider.GetPagesUrls(manga)
	// okay now we have all images links, so download all the images... if we have any
	if len(imageLinks) > 0 {
		tempDirectory, err := ioutil.TempDir("", manga.Title)
		if err != nil {
			dialog.ShowError(err, win)
			return
		} else {
			pageNumber := 0
			nbPages := len(imageLinks)
			for {
				//nbJobs := runtime.NumCPU() - 1
				nbJobs := 4
				if (pageNumber + nbJobs) > nbPages {
					nbJobs = nbPages - pageNumber
				}
				jobs := make(chan downloadWorkerJob, nbJobs)
				results := make(chan downloadWorkerResult, nbJobs)
				wg := sync.WaitGroup{}
				for i := 0; i < nbJobs; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						downloadWorker(jobs, results)
					}()
				}
				go func() {
					defer close(results)
					wg.Wait()
				}()
				go func() {
					defer close(jobs)
					for i := 0; i < nbJobs; i++ {
						job := downloadWorkerJob{
							path: tempDirectory,
							page: pageNumber,
							url:  imageLinks[pageNumber],
						}
						jobs <- job
						pageNumber++
					}
				}()
				for res := range results {
					prog.SetValue(float64(pageNumber) / float64(nbPages))
					if res.error != nil {
						msg := errors.New(fmt.Sprintf("Status code %d when downloading page %d from url:\n%s\nthe error is: %s\nclick OK to continue...", res.statusCode, pageNumber, res.url, res.error))
						dialog.ShowError(msg, win)
						break
					}
				}
				if pageNumber >= nbPages {
					break
				}
			}
			// and now create the new cbz from that temporary directory
			err = createCBZ(manga.Path, tempDirectory, manga.Title, manga.LastChapter)
			if err != nil {
				dialog.ShowError(err, win)
				return
			} else {
				prog.SetValue(1)
				// update history
				lastChapterIndex := -1
				for i := 0; i < len(manga.Chapters); i++ {
					if manga.Chapters[i] > manga.LastChapter {
						lastChapterIndex = i
						break
					}
				}
				prog.Hide()
				if lastChapterIndex >= 0 {
					manga.LastChapter = manga.Chapters[lastChapterIndex]
					*globalConfig = settings.UpdateHistory(*globalConfig, manga)
					if manga.LastChapter <= provider.CheckLastChapter(manga) {
						showDownloadChapters(app, win, manga)
					}
				}
				//updateLibrary(app, win)
				prog.Show()
			}
		}
	}
}
