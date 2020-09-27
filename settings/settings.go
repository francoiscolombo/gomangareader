package settings

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Settings is the structure that allowed to store the default configuration and the download history for all the mangas
type Settings struct {
	Config  Config  `json:"config"`
	History History `json:"history"`
}

// Config only store the default configuration, like output path and the global site provider
type Config struct {
	LibraryPath string `json:"library_path"`
	Provider    string `json:"provider"`
}

// History is the manga download history, so it's an array of all the mangas downloaded
type History struct {
	Titles []Manga `json:"titles"`
}

// Manga keep the download history for every mangas that we are subscribing
type Manga struct {
	Title         string `json:"title"`
	LastChapter   int    `json:"last_chapter"`
	CoverPath     string `json:"cover_path"`
	Path          string `json:"path"`
	CoverUrl      string `json:"cover_url"`
	Name          string `json:"name"`
	AlternateName string `json:"alternate_name"`
	YearOfRelease string `json:"year_of_release"`
	Status        string `json:"status"`
	Author        string `json:"author"`
	Artist        string `json:"artist"`
	ReadingDirection string `json:"reading_direction"`
	Description      string `json:"description"`
}

func getSettingsPath() string {
	user, err := user.Current()
	if err != nil {
		fmt.Printf("Error when trying to get current user: %s\n", err)
		os.Exit(1)
	}
	return user.HomeDir + "/.gomangareader.json"
}

/*
IsSettingsExisting allows to check if the settings file already exists or no
*/
func IsSettingsExisting() bool {
	if _, err := os.Stat(getSettingsPath()); !os.IsNotExist(err) {
		return true
	}
	return false
}

/*
WriteDefaultSettings write the default settings
*/
func WriteDefaultSettings() {
	user, err := user.Current()
	if err != nil {
		fmt.Printf("Error when trying to get current user: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Hello %s ! You don't have any settings yet. I can see that your homedir is %s, I will use it if you don't mind.\n", user.Name, user.HomeDir)
	settings := Settings{
		Config{
			LibraryPath: fmt.Sprintf("%s/mangas", user.HomeDir),
			Provider:    "mangareader.net",
		},
		History{
			Titles: []Manga{},
		},
	}
	file, _ := json.MarshalIndent(settings, "", " ")
	_ = ioutil.WriteFile(getSettingsPath(), file, 0644)
}

/*
ReadSettings read the settings file
*/
func ReadSettings() (settings Settings) {

	// Open our jsonFile
	settingsPath := getSettingsPath()
	fmt.Printf("Loading settings from %s...\n", settingsPath)
	jsonFile, err := os.Open(settingsPath)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Printf("Error when trying to open settings file: %s\n", err)
	}

	fmt.Println("Successfully Opened settings.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'settings' which we defined above
	json.Unmarshal(byteValue, &settings)

	return
}

/*
WriteSettings write a settings file. used to change the default config or add manga to history download
*/
func WriteSettings(settings Settings) {
	file, _ := json.MarshalIndent(settings, "", " ")
	_ = ioutil.WriteFile(getSettingsPath(), file, 0644)
}

/*
UpdateHistory register the last chapter downloaded for a manga, and the last provider used
*/
func UpdateHistory(cfg Settings, manga string, chapter int) (newSettings Settings) {
	if chapter < 0 {
		chapter = 1
	}
	var titles []Manga
	for _, title := range cfg.History.Titles {
		if title.Title != manga {
			titles = append(titles, Manga{
				Title:       title.Title,
				LastChapter: title.LastChapter,
			})
		}
	}
	titles = append(titles, Manga{
		Title:       manga,
		LastChapter: chapter,
	})
	newSettings = Settings{
		Config{
			LibraryPath: cfg.Config.LibraryPath,
			Provider:    cfg.Config.Provider,
		},
		History{
			Titles: titles,
		},
	}
	WriteSettings(newSettings)
	fmt.Println("History updated.")
	return
}

/*
SearchLastChapter send the last chapter in the history for a manga, or 1 if no history exists yet
*/
func SearchLastChapter(settings Settings, manga string) (lastChapter int) {
	lastChapter = 1
	for _, title := range settings.History.Titles {
		if title.Title == manga {
			lastChapter = title.LastChapter
			break
		}
	}
	return
}

/*
UpdateMetaData will retrieve all metadata directly from mangareader.net (no other provider accepted here)
it will fill all the settings metadata but also download the cover of the serie, and generate a thumbnail
for every cbz existing (it will be the first page of the cbz)
all the additional graphics (serie cover, thumbnails) are generated on the following directory:
<manga path>/.meta/<manga title>-xxxx
the serie cover will be named "cover.jpg" and the thumbnail 000 - nb of chapter.jpg
*/
func UpdateMetaData(config Settings) {
	// create metadata directory if it does not exists yet
	metadataPath := filepath.FromSlash(fmt.Sprintf("%s/.metadata", config.Config.LibraryPath))
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		err := os.MkdirAll(metadataPath,os.ModePerm)
		if err != nil {
			log.Fatalf("Error while trying to create directory %s: %s",metadataPath,err)
		}
	}
	var mangaUpdatedList []Manga
	for _, manga := range config.History.Titles {
		fmt.Printf("- Refresh metadatas for manga %s now... ",manga.Title)
		// okay at this point we are sure to have the directory
		// access detail data from mangareader.net only
		// this is working only with this website
		// cover image: present in div.d38
		// description: d46.p
		// details: table.d41
		// - first line: name
		// - second line: alternate name
		// - third line: year of release
		// - forth line: status
		// - fifth line: author
		// - sixth line: artist
		// - seventh line: reading direction
		// - eighth line: genre (a href with class a.d42)
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
		// extract cover url
		var coverUrl string
		doc.Find(".d38").Each(func(i int, div *goquery.Selection) {
			div.Find("img").Each(func(i int, img *goquery.Selection) {
				v, _ := img.Attr("src")
				if strings.HasPrefix(v,"http") {
					coverUrl = v
				} else {
					coverUrl = fmt.Sprintf("https:%s",v)
				}
				return
			})
		})
		// extract description
		var description string
		doc.Find(".d46").Each(func(i int, div *goquery.Selection) {
			div.Find("p").Each(func(i int, p *goquery.Selection) {
				description = p.Text()
				return
			})
		})
		// parse table.d41
		var name string
		var alternateName string
		var yearOfRelease string
		var status string
		var author string
		var artist string
		var readingDirection string
		doc.Find(".d41").Each(func(i int, table *goquery.Selection) {
			table.Find("tr").Each(func(ir int, row *goquery.Selection) {
				row.Find("td").Each(func(ic int, cell *goquery.Selection) {
					if ic == 1 {
						if ir == 0 {
							name = cell.Text()
						} else if ir == 1 {
							alternateName = cell.Text()
						} else if ir == 2 {
							yearOfRelease = cell.Text()
						} else if ir == 3 {
							status = cell.Text()
						} else if ir == 4 {
							author = cell.Text()
						} else if ir == 5 {
							artist = cell.Text()
						} else if ir == 6 {
							readingDirection = cell.Text()
						}
					}
					return
				})
				return
			})
		})
		coverPath := filepath.FromSlash(fmt.Sprintf("%s/%s-cover.jpg",metadataPath,manga.Title))
		// create structure with details to keep
		newManga := Manga{
			Title:         manga.Title,
			LastChapter:   manga.LastChapter,
			CoverPath:     coverPath,
			CoverUrl:      coverUrl,
			Path:          filepath.FromSlash(fmt.Sprintf("%s/%s", config.Config.LibraryPath,manga.Title)),
			Name:          name,
			AlternateName: alternateName,
			YearOfRelease: yearOfRelease,
			Status:        status,
			Author:        author,
			Artist:        artist,
			ReadingDirection: readingDirection,
			Description: description,
		}
		mangaUpdatedList = append(mangaUpdatedList, newManga)
		// download cover picture (if needed)
		downloadCover(newManga)
		// and generate thumbnails (if needed)
		extractFirstPages(config.Config.LibraryPath,newManga)
		fmt.Println(" completed.")
	}
	// okay we have updated the metadata, now we can save the config
	newSettings := Settings{
		Config{
			LibraryPath: config.Config.LibraryPath,
			Provider:    config.Config.Provider,
		},
		History{
			Titles: mangaUpdatedList,
		},
	}
	WriteSettings(newSettings)
	fmt.Println("> Settings updated.")
}

/*
DownloadCover simply download a cover for a manga title
*/
func downloadCover(manga Manga) {
	// download only if does not exists
	if _, err := os.Stat(manga.CoverPath); os.IsNotExist(err) {
		fmt.Printf("- %s does not exists yet, we have to download it....",manga.CoverPath)
		req, _ := http.NewRequest("GET", manga.CoverUrl, nil)
		req.Header.Add("cache-control", "no-cache")
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			// handle error
			log.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			log.Printf("status code error while trying to extract images from %s: %d %s", manga.CoverUrl, res.StatusCode, res.Status)
		} else {
			//open a file for writing
			file, err := os.Create(manga.CoverPath)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			// Use io.Copy to just dump the response body to the file. This supports huge files
			_, err = io.Copy(file, res.Body)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("- new cover downloaded")
		}
	}
}

/*
ExtractFirstPage allows to extract the first page of a cbz archive to generate a thumbnail
*/
func extractFirstPages(globalPath string, manga Manga) {
	for i := 1; i < manga.LastChapter; i++ {
		cbzArchive := filepath.FromSlash(fmt.Sprintf("%s/%s/%s-%03d.cbz",globalPath,manga.Title,manga.Title,i))
		cbzThumbnail := filepath.FromSlash(fmt.Sprintf("%s/.metadata/%s-%03d.jpg",globalPath,manga.Title,i))
		if _, err := os.Stat(cbzThumbnail); os.IsNotExist(err) {
			r, err := zip.OpenReader(cbzArchive)
			if err != nil {
				fmt.Printf("[Skipped] Error while opening archive %s: %s",cbzArchive,err)
			} else {
				f := r.File[0]
				outFile, err := os.OpenFile(cbzThumbnail, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					log.Fatalf("Error while opening thumbnail %s: %s",cbzThumbnail,err)
				}
				rc, err := f.Open()
				if err != nil {
					log.Fatalf("Error while opening first page of %s: %s",cbzThumbnail,err)
				}
				_, err = io.Copy(outFile, rc)
				_ = outFile.Close()
				_ = rc.Close()
				if err != nil {
					log.Fatalf("Error while saving %s: %s",cbzThumbnail,err)
				}
				err = r.Close()
				if err != nil {
					log.Fatalf("Error while closing archive %s: %s",cbzArchive,err)
				}
			}
		}
	}
}
