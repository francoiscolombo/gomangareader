package main

import (
	"fmt"

	"github.com/francoiscolombo/gomangareader/settings"
	"github.com/francoiscolombo/gomangareader/widget"
)

const (
	versionNumber = "1.0"
	versionName   = "Another Dimension"
)

func main() {

	fmt.Println("\nWelcome on gomangareader")
	fmt.Println("------------------------\n")

	fmt.Printf("version %s (%s)\n", versionNumber, versionName)

	if settings.IsSettingsExisting() == false {
		settings.WriteDefaultSettings()
	}

	cfg := settings.ReadSettings()

	fmt.Println("- Settings loaded.")
	fmt.Printf("  > Library path is %s\n  > Default provider is %s\n\n", cfg.Config.LibraryPath, cfg.Config.Provider)

	settings.UpdateMetaData(cfg)

	widget.ShowLibrary(&cfg)

}
