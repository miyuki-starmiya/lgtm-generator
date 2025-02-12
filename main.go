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
	var (
		lgtmPath string
		width    int
	)

	// Parse flags
	basePath := flag.String("input", "", "Input image path (gif, png, jpg, etc.)")
	// lgtmFlag := flag.String("lgtm", "assets/lgtm.png", "Path to LGTM overlay image")
	targetWidthFlag := flag.Int("width", 320, "Target width for resizing")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	inputPath := fmt.Sprintf("%s/input/%s", wd, *basePath)
	outputPath := fmt.Sprintf("output/%d_lgtm_%s", *targetWidthFlag, filepath.Base(*basePath))

	if inputPath == "" {
		log.Fatal("You must specify --input")
	}
	if *targetWidthFlag == 320 {
		lgtmPath = "assets/320_lgtm.png"
		width = 320
	} else {
		lgtmPath = fmt.Sprintf("assets/%d_lgtm.png", *targetWidthFlag)
		width = *targetWidthFlag
	}
	if _, err := os.Stat(lgtmPath); os.IsNotExist(err) {
		log.Fatalf("LGTM image does not exist: %s", lgtmPath)
	}

	// Detect GIF by file extension
	ext := strings.ToLower(filepath.Ext(inputPath))
	switch ext {
	case ".gif":
		err := ProcessAnimatedGIF(inputPath, outputPath, lgtmPath, width)
		if err != nil {
			log.Fatalf("GIF processing failed: %v", err)
		}
		fmt.Println("Animated GIF generated successfully.")
	default:
		err := ProcessStaticImage(inputPath, outputPath, lgtmPath, width)
		if err != nil {
			log.Fatalf("Static processing failed: %v", err)
		}
		fmt.Println("Static image generated successfully.")
	}
}
