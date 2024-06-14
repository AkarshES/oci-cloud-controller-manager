package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("expected 'cleanup-images' or 'validate-mappings' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "cleanup-images":
		cleanupUnusedImages()
	case "validate-mappings":
		validateMappings()
	}
}

func cleanupUnusedImages() {
	// Read the CSV file
	manifestCsvFile, err := os.Open("MANIFEST.csv")
	if err != nil {
		fmt.Println("Error opening MANIFEST.csv:", err)
		return
	}
	defer manifestCsvFile.Close()

	manifestCsvReader := csv.NewReader(manifestCsvFile)
	manifestCsvReader.Comma = ','

	manifestCsvRecords, err := manifestCsvReader.ReadAll()
	if err != nil {
		fmt.Println("Error reading contents of MANIFEST.csv:", err)
		return
	}

	// Create a map of image tags
	imageTags := make(map[string]bool)
	for _, record := range manifestCsvRecords {
		imageTags[strings.TrimSpace(record[1])] = false
	}

	checkImageUsageInFile("shepherd/limits/shared_modules/properties_values/default_values.tf", imageTags)
	checkImageUsageInFile("shepherd/limits/shared_modules/properties_values/locals.tf", imageTags)

	manifestCsvFile, err = os.Open("MANIFEST.csv")
	if err != nil {
		fmt.Println("Error opening MANIFEST.csv file:", err)
		return
	}
	defer manifestCsvFile.Close()

	manifestArchiveFile, err := os.OpenFile("MANIFEST_ARCHIVE.csv", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening  MANIFEST_ARCHIVE.csv file:", err)
		return
	}
	defer manifestArchiveFile.Close()

	scanner := bufio.NewScanner(manifestCsvFile)
	manifestArchiveWriter := bufio.NewWriter(manifestArchiveFile)

	var recordsToKeep []string
	for scanner.Scan() {

		line := scanner.Text()
		imageNameTagRecord := strings.SplitN(line, ",", 2)

		if len(imageNameTagRecord) != 2 {
			fmt.Println("Invalid record:", line)
			continue
		}
		imageTag := strings.TrimSpace(imageNameTagRecord[1])
		if keep, exists := imageTags[imageTag]; exists && keep {
			recordsToKeep = append(recordsToKeep, line)
		} else {
			fmt.Fprintln(manifestArchiveWriter, line)
		}
	}
	manifestArchiveWriter.Flush()
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return
	}

	outputFilePath := "MANIFEST.csv" // Replace with your desired output file path
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, record := range recordsToKeep {
		fmt.Fprintln(writer, record)
	}
	writer.Flush()
}

// Check the image tags against the text files
func checkImageUsageInFile(filename string, imageTags map[string]bool) {
	content, err := os.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}

	// Convert []byte to string and print to screen
	text := string(content)
	for key := range imageTags {
		if strings.Contains(text, key) {
			imageTags[key] = true
		}
	}
}

func validateMappings() {

	oldImagesNotPresentInManifest := map[string]struct{}{
		"oke-1.23-b8a8dee-140":           {},
		"oke-multiarch-1.23-16022dc-95":  {},
		"oke-multiarch-1.16-520cc1d-11":  {},
		"oke-1.16-520cc1d-11":            {},
		"oke-1.17-40e9a7a-13":            {},
		"oke-1.22-aa9948e-192":           {},
		"oke-1.23-16022dc-95":            {},
		"oke-1.22-9f30fe0-220":           {},
		"oke-multiarch-1.22-d0bafe8-232": {},
		"oke-multiarch-1.22-aa9948e-192": {},
	}

	// Read the CSV file
	manifestCsvFile, err := os.Open("MANIFEST.csv")
	if err != nil {
		fmt.Println("Error opening MANIFEST.csv:", err)
		return
	}
	defer manifestCsvFile.Close()

	manifestCsvReader := csv.NewReader(manifestCsvFile)
	manifestCsvReader.Comma = ','

	manifestCsvRecords, err := manifestCsvReader.ReadAll()
	if err != nil {
		fmt.Println("Error reading contents of MANIFEST.csv:", err)
		return
	}

	// Create a map of image tags
	imageTags := make(map[string]struct{})
	for _, record := range manifestCsvRecords {
		imageTags[strings.TrimSpace(record[1])] = struct{}{}
	}

	validImages := make(map[string]struct{})
	invalidImages := make(map[string]struct{})

	err = validateImagesInMappingsWithManifest("shepherd/limits/shared_modules/properties_values/default_values.tf",
		imageTags, oldImagesNotPresentInManifest, validImages, invalidImages)
	if err != nil {
		fmt.Println("Error validating mappings:", err)
		return
	}
	err = validateImagesInMappingsWithManifest("shepherd/limits/shared_modules/properties_values/locals.tf",
		imageTags, oldImagesNotPresentInManifest, validImages, invalidImages)
	if err != nil {
		fmt.Println("Error validating mappings:", err)
		return
	}
	if len(validImages) > 0 {
		fmt.Println("Below Image tags used in mappings are valid : ")
		for key := range validImages {
			fmt.Println(key)
		}
	}

	if len(invalidImages) > 0 {
		fmt.Println("\n\nBelow Image tags used in mappings are Invalid : ")
		for key := range invalidImages {
			fmt.Println(key)
		}
	} else {
		fmt.Println("\n\nAll the images used in mappings are present in MANIFEST.csv")
	}

}

func validateImagesInMappingsWithManifest(mappingFile string, imageTags map[string]struct{}, oldImagesNotPresentInManifest map[string]struct{},
	validImages map[string]struct{}, invalidImages map[string]struct{}) error {

	file, err := os.Open(mappingFile)

	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	re := regexp.MustCompile(`"([\w\d\-\.]+)@sha256:[\w\d]+"`)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			_, exists := imageTags[strings.TrimSpace(matches[1])]
			_, existsInOldImages := oldImagesNotPresentInManifest[matches[1]]
			if exists || existsInOldImages {
				validImages[strings.TrimSpace(matches[1])] = struct{}{}
			} else {
				invalidImages[strings.TrimSpace(matches[1])] = struct{}{}
			}
		}
	}
	return nil
}
