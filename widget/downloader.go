package widget

import (
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"github.com/francoiscolombo/gomangareader/settings"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Downloader struct {
	widget.BaseWidget
	Application     fyne.App
	SelectedManga   *settings.Manga
	DownloadChapter int
	CurrentPage     int
	TotalPages      int
}

func NewDownloader(app fyne.App, manga *settings.Manga, chapter int) *Downloader {
	d := &Downloader{
		Application:     app,
		SelectedManga:   manga,
		DownloadChapter: chapter,
		CurrentPage:     0,
		TotalPages:      0,
	}
	d.ExtendBaseWidget(d)
	return d
}

// MinSize returns the size that this widget should not shrink below
func (d *Downloader) MinSize() fyne.Size {
	d.ExtendBaseWidget(d)
	return d.BaseWidget.MinSize()
}

func (d *Downloader) CreateRenderer() fyne.WidgetRenderer {
	d.ExtendBaseWidget(d)
	var page *canvas.Image
	if d.SelectedManga.CoverPath != "" {
		page = canvas.NewImageFromFile(d.SelectedManga.CoverPath)
		page.FillMode = canvas.ImageFillContain
	}

	label := canvas.NewText("Please wait, downloading pages now...", theme.TextColor())
	label.TextSize = 12
	label.Alignment = fyne.TextAlignCenter

	progress := widget.NewProgressBar()

	description := widget.NewLabel(fmt.Sprintf(
		"%s\nChapter %03.1f - page %d / %d",
		d.SelectedManga.Name,
		d.SelectedManga.Chapters[d.DownloadChapter],
		d.CurrentPage,
		d.TotalPages,
	))
	description.Wrapping = fyne.TextWrapWord
	description.Alignment = fyne.TextAlignCenter

	download := widget.NewButtonWithIcon("Download chapters...", theme.DownloadIcon(), func() {
		d.ChapterDownloader()
	})

	bg := canvas.NewRectangle(theme.ButtonColor())

	dr := &DownloaderRenderer{
		bg:          bg,
		download:    download,
		page:        page,
		label:       label,
		progress:    progress,
		description: description,
		layout:      nil,
		downloader:  d,
	}

	return dr
}

func (d *Downloader) ChapterDownloader() {
	provider := settings.MangaReader{}
	imageLinks := provider.GetPagesUrls(*d.SelectedManga)
	// okay now we have all images links, so download all the images... if we have any
	if len(imageLinks) > 0 {
		tempDirectory, err := ioutil.TempDir("", d.SelectedManga.Title)
		if err == nil {
			d.CurrentPage = 0
			d.TotalPages = len(imageLinks)
			for {
				nbJobs := globalConfig.Config.NbWorkers
				if (d.CurrentPage + nbJobs) > d.TotalPages {
					nbJobs = d.TotalPages - d.CurrentPage
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
							path:       tempDirectory,
							page:       d.CurrentPage,
							url:        imageLinks[d.CurrentPage],
							downloader: d,
						}
						jobs <- job
						d.CurrentPage++
					}
				}()
				for res := range results {
					if res.error != nil {
						log.Printf("Status code %d when downloading page %d from url:\n%s\nthe error is: %s", res.statusCode, d.CurrentPage, res.url, res.error)
						break
					}
				}
				if d.CurrentPage >= d.TotalPages {
					break
				}
			}
			// and now create the new cbz from that temporary directory
			err = createCBZ(d.SelectedManga.Path, tempDirectory, d.SelectedManga.Title, d.SelectedManga.LastChapter)
			if err != nil {
				log.Printf("Error when trying to create chapter %03.1f of %s from %s\nthe error is: %s", d.SelectedManga.LastChapter, d.SelectedManga.Title, tempDirectory, err)
				return
			}
			// update history
			lastChapterIndex := -1
			for i := 0; i < len(d.SelectedManga.Chapters); i++ {
				if d.SelectedManga.Chapters[i] > d.SelectedManga.LastChapter {
					lastChapterIndex = i
					break
				}
			}
			if lastChapterIndex >= 0 {
				d.SelectedManga.LastChapter = d.SelectedManga.Chapters[lastChapterIndex]
				*globalConfig = settings.UpdateHistory(*globalConfig, *d.SelectedManga)
				if d.SelectedManga.LastChapter <= provider.CheckLastChapter(*d.SelectedManga) {
					d.CurrentPage = 0
					d.TotalPages = 1
					d.DownloadChapter = lastChapterIndex
					d.Refresh()
				}
				err = extractFirstPages(globalConfig.Config.LibraryPath, *d.SelectedManga)
				if err != nil {
					log.Printf("Error happened while extracting first page for %s\n%s", d.SelectedManga.Name, err)
				}
			}
		}
	}
}

type DownloaderRenderer struct {
	bg          *canvas.Rectangle
	page        *canvas.Image
	label       *canvas.Text
	progress    *widget.ProgressBar
	description *widget.Label
	download    *widget.Button
	layout      fyne.Layout
	downloader  *Downloader
}

func (d *DownloaderRenderer) BackgroundColor() color.Color {
	return theme.BackgroundColor()
}

func (d *DownloaderRenderer) Destroy() {
	d.bg = nil
	d.download = nil
	d.page = nil
	d.progress = nil
	d.label = nil
	d.description = nil
	d.layout = nil
	d.downloader = nil
}

func (d *DownloaderRenderer) MinSize() fyne.Size {
	return fyne.NewSize(globalConfig.Config.ThumbnailWidth*6+theme.Padding()*2, globalConfig.Config.ThumbMiniHeight+theme.Padding()*2)
}

func (d *DownloaderRenderer) Objects() []fyne.CanvasObject {
	var objects []fyne.CanvasObject
	objects = append(objects, d.bg)
	objects = append(objects, d.download)
	objects = append(objects, d.page)
	objects = append(objects, d.progress)
	objects = append(objects, d.label)
	objects = append(objects, d.description)
	return objects
}

func (d *DownloaderRenderer) Layout(size fyne.Size) {
	p := theme.Padding()
	dx := p
	dy := p
	txtHeight := 20

	d.page.Resize(fyne.NewSize(globalConfig.Config.ThumbMiniWidth, globalConfig.Config.ThumbMiniHeight))
	d.page.Move(fyne.NewPos(dx, dy))
	dx = dx + globalConfig.Config.ThumbMiniWidth + p

	d.download.Resize(fyne.NewSize(200, globalConfig.Config.ThumbMiniHeight/2-p))
	d.download.Move(fyne.NewPos(globalConfig.Config.ThumbnailWidth*6-200-p, dy))

	d.label.Resize(fyne.NewSize(globalConfig.Config.ThumbnailWidth*5-200, txtHeight))
	d.label.Move(fyne.NewPos(dx, dy))
	dy = dy + txtHeight + p

	d.progress.Resize(fyne.NewSize(globalConfig.Config.ThumbnailWidth*5-200, txtHeight))
	d.progress.Move(fyne.NewPos(dx, dy))
	dy = dy + txtHeight + p

	d.description.Resize(fyne.NewSize(globalConfig.Config.ThumbnailWidth*5, txtHeight*2))
	d.description.Move(fyne.NewPos(dx, dy))
}

func (d *DownloaderRenderer) Refresh() {
	value := float64(d.downloader.CurrentPage) / float64(d.downloader.TotalPages)
	d.progress.SetValue(value)
	d.bg.Refresh()
	d.page.Refresh()
	d.label.Refresh()
	d.progress.Refresh()
	d.description = widget.NewLabel(fmt.Sprintf(
		"%s\nChapter %03.1f - page %d / %d",
		d.downloader.SelectedManga.Name,
		d.downloader.SelectedManga.Chapters[d.downloader.DownloadChapter],
		d.downloader.CurrentPage,
		d.downloader.TotalPages,
	))
	d.description.Refresh()
}

type downloadWorkerJob struct {
	path       string
	page       int
	url        string
	downloader *Downloader
}

type downloadWorkerResult struct {
	error      error
	url        string
	statusCode int
}

func downloadWorker(jobs <-chan downloadWorkerJob, results chan<- downloadWorkerResult) {
	for workerJob := range jobs {
		//log.Printf("Start download page %d from url %s in temp directory %s\n",workerJob.page,workerJob.url,workerJob.path)
		status, err := downloadImage(workerJob.path, workerJob.page, workerJob.url, workerJob.downloader)
		//log.Printf("Page %d downloaded with status code %d\n",workerJob.page,status)
		res := downloadWorkerResult{error: err, url: workerJob.url, statusCode: status}
		results <- res
	}
}

func downloadImage(path string, page int, url string, downloader *Downloader) (int, error) {
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

	downloader.Refresh()

	return resp.StatusCode, err
}
