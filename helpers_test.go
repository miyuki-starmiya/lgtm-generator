package main

import (
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

var (
	srcColor  = color.NRGBA{R: 0, G: 0, B: 255, A: 255}
	lgtmColor = color.NRGBA{R: 255, G: 0, B: 0, A: 255}
)

// setupWorkspace creates input/, output/ and assets/320_lgtm.png in a temp dir
// and changes into it, because the CLI uses paths relative to the working dir.
func setupWorkspace(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{"input", "output", "assets"} {
		if err := os.Mkdir(filepath.Join(dir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writePNG(t, filepath.Join(dir, "assets", "320_lgtm.png"), 100, 40, lgtmColor)

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})
	return dir
}

func solid(w, h int, c color.Color) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func writePNG(t *testing.T, path string, w, h int, c color.Color) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, solid(w, h, c)); err != nil {
		t.Fatal(err)
	}
}

// writeGIF writes an animated GIF with the given number of full-size frames.
func writeGIF(t *testing.T, path string, w, h, frames int) {
	t.Helper()
	pal := color.Palette{color.Black, color.White, srcColor}
	g := &gif.GIF{Config: image.Config{Width: w, Height: h, ColorModel: pal}}
	for i := 0; i < frames; i++ {
		frame := image.NewPaletted(image.Rect(0, 0, w, h), pal)
		for j := range frame.Pix {
			frame.Pix[j] = uint8(i % len(pal))
		}
		g.Image = append(g.Image, frame)
		g.Delay = append(g.Delay, 10*(i+1))
		g.Disposal = append(g.Disposal, gif.DisposalNone)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := gif.EncodeAll(f, g); err != nil {
		t.Fatal(err)
	}
}

func touch(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
}
