package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func writeToFile(filePath string, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return err
	}

	return nil
}

func processFile(filePath string, wg *sync.WaitGroup, resultChan chan<- string, progress *int64, totalFiles int) {
	defer wg.Done()

	sha256, err := calculateSHA256(filePath)
	if err != nil {
		fmt.Printf("Error calculating SHA256 for file %s: %s\n", filePath, err)
		resultChan <- ""
		return
	}

	result := fmt.Sprintf("| %s | %s |\n", filepath.Base(filePath), sha256)
	resultChan <- result

	// Update progress
	atomic.AddInt64(progress, 1)
	currentProgress := atomic.LoadInt64(progress)
	fmt.Printf("Progress: %d/%d\n", currentProgress, int64(totalFiles))
}

func main() {
	pathPtr := flag.String("p", "", "Target folder path")
	flag.Parse()

	if *pathPtr == "" {
		fmt.Println("Please provide a target folder path using -p flag")
		return
	}

	files, err := ioutil.ReadDir(*pathPtr)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return
	}

	readmeFilePath := filepath.Join(*pathPtr, "SHA256.md")

	table := "| Filename | SHA256 |\n| --- | --- |\n"

	var wg sync.WaitGroup
	resultChan := make(chan string)
	progress := int64(0)
	totalFiles := len(files)

	startTime := time.Now()

	for _, file := range files {
		if !file.IsDir() && !strings.EqualFold(file.Name(), "SHA256.md") {
			filePath := filepath.Join(*pathPtr, file.Name())
			wg.Add(1)
			go processFile(filePath, &wg, resultChan, &progress, totalFiles)
		}
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for result := range resultChan {
		if result != "" {
			table += result
		}
	}

	err = writeToFile(readmeFilePath, table)
	if err != nil {
		fmt.Println("Error writing to SHA256.md:", err)
		return
	}

	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime).Round(time.Second * 10).String()

	fmt.Println("Table has been written to SHA256.md successfully.")
	fmt.Println("Total time elapsed:", elapsedTime)
}
