// animated.go
package main

import (
	"fmt"
	"image"
	"image/gif"
	"os"

	"github.com/disintegration/imaging"
)

func ProcessAnimatedGIF(inputPath, outputPath, lgtmPath string, targetWidth int) error {
	// 1. Open the GIF file
	inFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("cannot open input file: %w", err)
	}
	defer inFile.Close()

	// Decode all frames
	g, err := gif.DecodeAll(inFile)
	if err != nil {
		return fmt.Errorf("gif.DecodeAll failed: %w", err)
	}

	// 2. Open LGTM image
	lgtmFile, err := os.Open(lgtmPath)
	if err != nil {
		return fmt.Errorf("failed to open LGTM image: %w", err)
	}
	defer lgtmFile.Close()

	lgtmImg, _, err := image.Decode(lgtmFile)
	if err != nil {
		return fmt.Errorf("failed to decode LGTM image: %w", err)
	}

	// 3. For each frame, "resize + overlay"
	for i := range g.Image {
		frame := g.Image[i] // *image.Paletted

		// Convert paletted frame to RGBA so we can manipulate it
		rgbaFrame := imaging.Clone(frame) // => *image.NRGBA
		// Our utility call:
		processed := ResizeAndOverlayLGTM(rgbaFrame, lgtmImg, targetWidth)

		// Now we must convert it back to paletted to store in g.Image[i]
		palettedFrame := ToPaletted(processed, frame)

		g.Image[i] = palettedFrame
		g.Config.Width = palettedFrame.Bounds().Dx()
		g.Config.Height = palettedFrame.Bounds().Dy()
	}

	// 4. Save new GIF
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer outFile.Close()

	if err := gif.EncodeAll(outFile, g); err != nil {
		return fmt.Errorf("gif.EncodeAll failed: %w", err)
	}

	return nil
}
