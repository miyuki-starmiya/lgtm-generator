// static.go
package main

import (
	"fmt"

	"github.com/disintegration/imaging"
)

func ProcessStaticImage(inputPath, outputPath, lgtmPath string, targetWidth int) error {
	// 1. Open input image (jpg, png, etc.)
	src, err := imaging.Open(inputPath)
	if err != nil {
		return fmt.Errorf("imaging.Open failed: %w", err)
	}

	// 2. Open LGTM image
	lgtmImg, err := imaging.Open(lgtmPath)
	if err != nil {
		return fmt.Errorf("failed to open LGTM image: %w", err)
	}

	// 3. Do the shared "resize + overlay" logic
	final := ResizeAndOverlayLGTM(src, lgtmImg, targetWidth)

	// 4. Save result
	if err := imaging.Save(final, outputPath); err != nil {
		return fmt.Errorf("failed to save static output: %w", err)
	}

	return nil
}
