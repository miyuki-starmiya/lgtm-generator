// main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Parse flags
	basePath := flag.String("input", "", "Input image path (gif, png, jpg, etc.)")
	lgtmPath := flag.String("lgtm", "assets/lgtm.png", "Path to LGTM overlay image")
	width := flag.Int("width", 500, "Target width for resizing")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	inputPath := fmt.Sprintf("%s/%s", wd, *basePath)
	outputPath := fmt.Sprintf("output/lgtm_%s", filepath.Base(*basePath))

	if inputPath == "" {
		log.Fatal("You must specify --input")
	}
	if _, err := os.Stat(*lgtmPath); os.IsNotExist(err) {
		log.Fatalf("LGTM image does not exist: %s", *lgtmPath)
	}

	// Detect GIF by file extension
	ext := strings.ToLower(filepath.Ext(inputPath))
	switch ext {
	case ".gif":
		err := ProcessAnimatedGIF(inputPath, outputPath, *lgtmPath, *width)
		if err != nil {
			log.Fatalf("GIF processing failed: %v", err)
		}
		fmt.Println("Animated GIF generated successfully.")
	default:
		err := ProcessStaticImage(inputPath, outputPath, *lgtmPath, *width)
		if err != nil {
			log.Fatalf("Static processing failed: %v", err)
		}
		fmt.Println("Static image generated successfully.")
	}
}
