module github.com/francoiscolombo/gomangareader/widget

go 1.18

require (
	fyne.io/fyne/v2 v2.2.4
	github.com/francoiscolombo/gomangareader/archive v0.0.0-00010101000000-000000000000
	github.com/francoiscolombo/gomangareader/settings v0.0.0-00010101000000-000000000000
)

require (
	github.com/PuerkitoBio/goquery v1.8.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/go-gl/gl v0.0.0-20211210172815-726fda9656d6 // indirect
	github.com/go-gl/glfw/v3.3/glfw v0.0.0-20211213063430-748e38ca8aec // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/goki/freetype v0.0.0-20181231101311-fa8a33aabaff // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/srwiley/oksvg v0.0.0-20200311192757-870daf9aa564 // indirect
	github.com/srwiley/rasterx v0.0.0-20200120212402-85cb7272f5e9 // indirect
	github.com/stretchr/testify v1.7.2 // indirect
	golang.org/x/image v0.0.0-20220601225756-64ec528b34cd // indirect
	golang.org/x/net v0.0.0-20210916014120-12bc252f5db8 // indirect
	golang.org/x/sys v0.0.0-20220412211240-33da011f77ad // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/francoiscolombo/gomangareader/archive => ../archive

replace github.com/francoiscolombo/gomangareader/settings => ../settings
