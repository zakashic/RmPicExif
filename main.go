package main

import (
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func removeJPGExif(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := file.Close(); cErr != nil && err == nil {
			err = cErr
		}
	}()

	img, err := jpeg.Decode(file)
	if err != nil {
		return err
	}

	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if oErr := file.Close(); oErr != nil && err == nil {
			err = oErr
		}
	}()

	options := jpeg.Options{Quality: 100}
	return jpeg.Encode(output, img, &options)
}

func removePNGExif(filePath string) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := file.Close(); cErr != nil && err == nil {
			err = cErr
		}
	}()

	img, err := png.Decode(file)
	if err != nil {
		return err
	}

	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if oErr := file.Close(); oErr != nil && err == nil {
			err = oErr
		}
	}()

	return png.Encode(output, img)
}

func worker(id int, jobs <-chan string, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()
	for path := range jobs {
		var err error
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg":
			err = removeJPGExif(path)
		case ".png":
			err = removePNGExif(path)
		}
		if err != nil {
			results <- fmt.Sprintf("Worker %d: Error processing file %s: %v", id, path, err)
		} else {
			results <- fmt.Sprintf("Worker %d: Successfully processed file: %s", id, path)
		}
	}
}

func processFolder(folderPath string) error {
	//return filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
	//	if err != nil {
	//		return err
	//	}
	//
	//	if !info.IsDir() {
	//		ext := strings.ToLower(filepath.Ext(path))
	//		var err error
	//		switch ext {
	//		case ".jpg", ".jpeg":
	//			err = removeJPGExif(path)
	//		case ".png":
	//			err = removePNGExif(path)
	//		}
	//		if err != nil {
	//			fmt.Printf("Error processing file %s: %v", path, err)
	//		} else {
	//			fmt.Printf("Successfully processed file: %s", path)
	//		}
	//	}
	//	return nil
	//})

	var wg sync.WaitGroup
	jobs := make(chan string, 1000)
	results := make(chan string, 1000)

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	var walkErr error
	go func() {
		defer close(jobs)
		walkErr = filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				ext := strings.ToLower(filepath.Ext(path))
				if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
					jobs <- path
				} else {
					fmt.Printf("Skipping non-target file: %s\n", path)
				}
			}
			return nil
		})
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var errs []string
	if walkErr != nil {
		errs = append(errs, fmt.Sprintf("Error walking the path: %v", walkErr))
	}
	for result := range results {
		if strings.HasPrefix(result, "Worker") && strings.Contains(result, "Error") {
			errs = append(errs, result)
		}
		//fmt.Println(result)
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors occurred:\n%s", strings.Join(errs, "\n"))
	}

	return nil
}

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
