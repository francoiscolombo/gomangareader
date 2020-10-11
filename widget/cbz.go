package widget

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func createCBZ(outputPath, pagesPath, title string, chapter int) error {
	// create output path
	err := os.MkdirAll(outputPath, os.ModePerm)
	if err != nil {
		return err
	}
	// List of Files to Zip
	var files []string
	outputCBZ := filepath.FromSlash(fmt.Sprintf("%s/%s-%03d.cbz", outputPath, title, chapter))
	//fmt.Printf("\ncreate %s ... ", outputCBZ)
	err = filepath.Walk(pagesPath, func(path string, info os.FileInfo, err error) error {
		src, err := os.Stat(path)
		if err != nil {
			// still does not exists? then something wrong, exit in panic mode.
			panic(err)
		}
		if !src.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	// create archive
	newZipFile, err := os.Create(outputCBZ)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		if err = addFileToZip(zipWriter, file); err != nil {
			return err
		} else {
			os.Remove(file)
		}
	}

	// remove temporary folder
	os.Remove(pagesPath)
	//fmt.Println("done")

	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	//header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
