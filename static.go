// static.go
package main

import (
	"fmt"
	"image"
	"os"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

func ProcessStaticImage(inputPath, outputPath, lgtmPath string, targetWidth int) error {
	src, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("imaging.Open failed: %w", err)
	}

	lgtmImg, err := imaging.Open(lgtmPath)
	if err != nil {
		return fmt.Errorf("failed to open LGTM image: %w", err)
	}

	final := ResizeAndOverlayLGTM(src, lgtmImg, targetWidth)

	return saveWebP(final, outputPath)
}

func saveWebP(img image.Image, outputPath string) error {
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	if err := webp.Encode(f, img, &webp.Options{Quality: 80}); err != nil {
		return fmt.Errorf("failed to encode WebP: %w", err)
	}
	return nil
}
