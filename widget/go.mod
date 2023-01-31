module github.com/francoiscolombo/gomangareader/widget

go 1.18

require (
	fyne.io/fyne/v2 v2.3.0
	github.com/francoiscolombo/gomangareader/archive v0.0.0-00010101000000-000000000000
	github.com/francoiscolombo/gomangareader/settings v0.0.0-00010101000000-000000000000
)

require (
	github.com/PuerkitoBio/goquery v1.8.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-gl/gl v0.0.0-20211210172815-726fda9656d6 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/image v0.0.0-20220601225756-64ec528b34cd // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/francoiscolombo/gomangareader/archive => ../archive

replace github.com/francoiscolombo/gomangareader/settings => ../settings
