// lgtm.go
package main

import (
	"image"
	"image/color"
	"image/draw"

	"github.com/disintegration/imaging"
)

// ResizeAndOverlayLGTM resizes src to targetWidth (keeping aspect ratio),
// then overlays lgtmImg in the center. Returns a new RGBA image.
func ResizeAndOverlayLGTM(src, lgtmImg image.Image, targetWidth int) image.Image {
	// 1. Convert src to RGBA (imaging.Clone will do this for us).
	rgbaSrc := imaging.Clone(src)

	// 2. Resize to the desired width, keep aspect ratio (height=0).
	resized := imaging.Resize(rgbaSrc, targetWidth, 0, imaging.Lanczos)

	// 3. Convert LGTM to RGBA if needed (Clone).
	rgbaLGTM := imaging.Clone(lgtmImg)

	// 4. Calculate offset for centering LGTM.
	bgBounds := resized.Bounds()
	lgtmBounds := rgbaLGTM.Bounds()

	offsetX := (bgBounds.Dx() - lgtmBounds.Dx()) / 2
	offsetY := (bgBounds.Dy() - lgtmBounds.Dy()) / 2

	// 5. Overlay the LGTM image on the resized image.
	finalImg := imaging.Overlay(resized, rgbaLGTM, image.Pt(offsetX, offsetY), 1.0)

	return finalImg
}

// ToPaletted converts an RGBA image to a *image.Paletted by reusing the palette
// from an existing paletted image (second parameter).
func ToPaletted(src image.Image, palettedFrame image.Image) *image.Paletted {
	// 1. Assert that palettedFrame is actually *image.Paletted
	frame, ok := palettedFrame.(*image.Paletted)
	if !ok {
		// fallback: define a simple palette or return nil/error
		// For example, we can define a 256-color (or smaller) palette manually
		simplePalette := color.Palette{
			color.Black, color.White, color.Transparent,
			// ... you can add more colors here ...
		}
		// create a new paletted image with fallback palette
		fallback := image.NewPaletted(src.Bounds(), simplePalette)
		draw.Draw(fallback, fallback.Bounds(), src, src.Bounds().Min, draw.Over)
		return fallback
	}

	cm := frame.ColorModel()      // cm is a color.Model
	pal, ok := cm.(color.Palette) // type-assert to color.Palette
	if !ok {
		// If the assertion fails, you need a fallback palette or handle the error.
		return nil // or return an error, etc.
	}

	// Now we have pal as a color.Palette
	palettedImg := image.NewPaletted(src.Bounds(), pal)
	draw.Draw(palettedImg, palettedImg.Bounds(), src, src.Bounds().Min, draw.Over)
	return palettedImg
}
