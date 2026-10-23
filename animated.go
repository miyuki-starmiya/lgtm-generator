// animated.go
package main

import (
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
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

	// 3. Composite each frame onto a full-size canvas (GIF frames are often
	// sub-rectangle deltas), then "resize + overlay" the whole canvas.
	canvasRect := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(canvasRect)
	outW, outH := 0, 0

	for i, frame := range g.Image {
		disposal := g.Disposal[i]

		// Remember the canvas for DisposalPrevious before drawing this frame.
		var prev *image.RGBA
		if disposal == gif.DisposalPrevious {
			prev = image.NewRGBA(canvasRect)
			draw.Draw(prev, canvasRect, canvas, image.Point{}, draw.Src)
		}

		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		processed := ResizeAndOverlayLGTM(canvas, lgtmImg, targetWidth)

		// Every output frame is full-size, so it starts at (0,0).
		palettedFrame := image.NewPaletted(processed.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(palettedFrame, palettedFrame.Bounds(), processed, processed.Bounds().Min)

		// Frame bounds must be read before the loop variable is replaced.
		frameBounds := frame.Bounds()
		g.Image[i] = palettedFrame
		g.Disposal[i] = gif.DisposalNone
		outW, outH = palettedFrame.Bounds().Dx(), palettedFrame.Bounds().Dy()

		// Apply this frame's original disposal before the next frame.
		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, frameBounds, image.Transparent, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			draw.Draw(canvas, canvasRect, prev, image.Point{}, draw.Src)
		}
	}
	g.Config.Width = outW
	g.Config.Height = outH

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
