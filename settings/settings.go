package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

func getSettingsPath() string {
	user, err := user.Current()
	if err != nil {
		log.Fatalf("Error when trying to get current user: %s\n", err)
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
		log.Fatalf("Error when trying to get current user: %s\n", err)
	}
	log.Printf("Hello %s ! You don't have any settings yet. I can see that your homedir is %s, I will use it if you don't mind.\n", user.Name, user.HomeDir)
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
		log.Fatalf("Error when trying to open settings file: %s\n", err)
	}

	log.Println("Successfully Opened settings.json")
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
	log.Println("History updated.")
	return
}
