module github.com/francoiscolombo/gomangareader/settings

go 1.18

require (
	fyne.io/fyne v1.4.3
	github.com/PuerkitoBio/goquery v1.8.0
)

replace github.com/francoiscolombo/gomangareader/archive => ../archive

replace github.com/francoiscolombo/gomangareader/widget => ../widget
