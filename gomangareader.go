package main

import (
	"fmt"

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

	widget.ShowLibrary()

}
