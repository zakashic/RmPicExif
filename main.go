package main

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a folder path.")
		return
	}

	folderPath := os.Args[1]

	if err := processFolder(folderPath); err != nil {
		fmt.Printf("Error processing folder: %v\n", err)
	} else {
		fmt.Println("Successfully removed EXIF data from all JPG and PNG files.")
	}
}

func processFolder(folderPath string) error {
	return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			var err error
			switch ext {
			case ".jpg", ".jpeg":
				err = removeJPGExif(path)
			case ".png":
				err = removePNGExif(path)
			}
			if err != nil {
				fmt.Printf("Error processing file %s: %v", path, err)
			} else {
				fmt.Printf("Successfully processed file: %s", path)
			}
		}
		return nil
	})
}

func removeJPGExif(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("Error closing file: %v", err)
		}
	}(file)

	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer output.Close()

	options := jpeg.Options{Quality: 100}
	return jpeg.Encode(output, img, &options)
}

func removePNGExif(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	img, err := png.Decode(file)
	if err != nil {
		return err
	}

	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer output.Close()

	return png.Encode(output, img)
}
