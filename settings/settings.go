package settings

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"sort"
)

func getSettingsPath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error when trying to get current usr: %s\n", err)
	}
	return usr.HomeDir + "/.gomangareader.json"
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
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Error when trying to get current usr: %s\n", err)
	}
	log.Printf("Hello %s ! You don't have any settings yet. I can see that your homedir is %s, I will use it if you don't mind.\n", usr.Name, usr.HomeDir)
	settings := Settings{
		Config{
			LibraryPath: fmt.Sprintf("%s/mangas", usr.HomeDir),
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
	//fmt.Printf("Loading settings from %s...\n", settingsPath)
	jsonFile, err := os.Open(settingsPath)
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Fatalf("Error when trying to open settings file: %s\n", err)
	}

	//log.Println("Successfully Opened settings.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			log.Printf("Something went wrong while trying to close the json file, error is %s", err)
		}
	}(jsonFile)

	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'settings' which we defined above
	err = json.Unmarshal(byteValue, &settings)
	if err != nil {
		return Settings{}
	}

	titles := settings.History.Titles
	sort.Slice(titles, func(i, j int) bool {
		return titles[i].Title < titles[j].Title
	})
	settings.History.Titles = titles

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
func UpdateHistory(cfg Settings, manga Manga) (newSettings Settings) {
	var titles []Manga
	for _, title := range cfg.History.Titles {
		if title.Title != manga.Title {
			titles = append(titles, title)
		}
	}
	titles = append(titles, manga)
	sort.Slice(titles, func(i, j int) bool {
		return titles[i].Title < titles[j].Title
	})
	newSettings = Settings{
		Config{
			LibraryPath: cfg.Config.LibraryPath,
		},
		History{
			Titles: titles,
		},
	}
	WriteSettings(newSettings)
	//log.Println("History updated.")
	return
}
