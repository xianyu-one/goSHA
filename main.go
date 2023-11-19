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

	for i, file := range files {
		if !file.IsDir() && !strings.EqualFold(file.Name(), "SHA256.md") {
			filePath := filepath.Join(*pathPtr, file.Name())
			sha256, err := calculateSHA256(filePath)
			if err != nil {
				fmt.Printf("Error calculating SHA256 for file %s: %s\n", file.Name(), err)
				continue
			}

			row := fmt.Sprintf("| %s | %s |\n", file.Name(), sha256)
			table += row
		}

		// Output progress
		fmt.Printf("Progress: %d/%d\n", i+1, len(files))
	}

	err = writeToFile(readmeFilePath, table)
	if err != nil {
		fmt.Println("Error writing to SHA:", err)
		return
	}

	fmt.Println("Table has been written to SHA256.md t successfully.")
}