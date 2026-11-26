package main

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/webp"
)

func near(a, b color.Color) bool {
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	d := func(x, y uint32) uint32 {
		if x > y {
			return x - y
		}
		return y - x
	}
	const tol = 0x2000 // lossy encoding and dithering shift colors slightly
	return d(ar, br) < tol && d(ag, bg) < tol && d(ab, bb) < tol
}

func TestResizeAndOverlayLGTM(t *testing.T) {
	src := solid(640, 480, srcColor)
	lgtm := solid(100, 40, lgtmColor)

	got := ResizeAndOverlayLGTM(src, lgtm, 320)

	if b := got.Bounds(); b.Dx() != 320 || b.Dy() != 240 {
		t.Fatalf("size = %dx%d, want 320x240", b.Dx(), b.Dy())
	}
	if c := got.At(160, 120); !near(c, lgtmColor) {
		t.Errorf("center = %v, want LGTM color", c)
	}
	if c := got.At(5, 5); !near(c, srcColor) {
		t.Errorf("corner = %v, want source color", c)
	}
}

func TestProcessStaticImage(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.png")
	lgtm := filepath.Join(dir, "lgtm.png")
	out := filepath.Join(dir, "out.webp")
	writePNG(t, in, 640, 480, srcColor)
	writePNG(t, lgtm, 100, 40, lgtmColor)

	if err := ProcessStaticImage(in, out, lgtm, 320); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte("RIFF")) || string(data[8:12]) != "WEBP" {
		t.Fatalf("output is not WebP: % x", data[:16])
	}
	// "VP8 " is the lossy bitstream; lossless would be "VP8L".
	if !bytes.Contains(data, []byte("VP8 ")) || bytes.Contains(data, []byte("VP8L")) {
		t.Errorf("output is not lossy WebP")
	}

	img, err := webp.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if b := img.Bounds(); b.Dx() != 320 || b.Dy() != 240 {
		t.Errorf("size = %dx%d, want 320x240", b.Dx(), b.Dy())
	}
	if c := img.At(160, 120); !near(c, lgtmColor) {
		t.Errorf("center = %v, want LGTM color", c)
	}
}

func TestProcessStaticImageErrors(t *testing.T) {
	dir := t.TempDir()
	lgtm := filepath.Join(dir, "lgtm.png")
	writePNG(t, lgtm, 100, 40, lgtmColor)
	in := filepath.Join(dir, "in.png")
	writePNG(t, in, 640, 480, srcColor)

	if err := ProcessStaticImage(filepath.Join(dir, "missing.png"), filepath.Join(dir, "o.webp"), lgtm, 320); err == nil {
		t.Error("missing input: want error")
	}
	if err := ProcessStaticImage(in, filepath.Join(dir, "o.webp"), filepath.Join(dir, "missing.png"), 320); err == nil {
		t.Error("missing LGTM image: want error")
	}
	if err := ProcessStaticImage(in, filepath.Join(dir, "no-such-dir", "o.webp"), lgtm, 320); err == nil {
		t.Error("unwritable output: want error")
	}
}

func TestProcessAnimatedGIF(t *testing.T) {
	dir := t.TempDir()
	in := filepath.Join(dir, "in.gif")
	lgtm := filepath.Join(dir, "lgtm.png")
	out := filepath.Join(dir, "out.gif")
	writeGIF(t, in, 640, 480, 3)
	writePNG(t, lgtm, 100, 40, lgtmColor)

	if err := ProcessAnimatedGIF(in, out, lgtm, 320); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	g, err := gif.DecodeAll(f)
	if err != nil {
		t.Fatal(err)
	}

	if len(g.Image) != 3 {
		t.Fatalf("frames = %d, want 3", len(g.Image))
	}
	if g.Config.Width != 320 || g.Config.Height != 240 {
		t.Errorf("config size = %dx%d, want 320x240", g.Config.Width, g.Config.Height)
	}
	for i, frame := range g.Image {
		if frame.Bounds() != image.Rect(0, 0, 320, 240) {
			t.Errorf("frame %d bounds = %v, want (0,0)-(320,240)", i, frame.Bounds())
		}
		if want := 10 * (i + 1); g.Delay[i] != want {
			t.Errorf("frame %d delay = %d, want %d", i, g.Delay[i], want)
		}
		if c := frame.At(160, 120); !near(c, lgtmColor) {
			t.Errorf("frame %d center = %v, want LGTM color", i, c)
		}
	}
}

func TestProcessAnimatedGIFErrors(t *testing.T) {
	dir := t.TempDir()
	lgtm := filepath.Join(dir, "lgtm.png")
	writePNG(t, lgtm, 100, 40, lgtmColor)
	notGIF := filepath.Join(dir, "not.gif")
	writePNG(t, notGIF, 10, 10, srcColor)

	if err := ProcessAnimatedGIF(filepath.Join(dir, "missing.gif"), filepath.Join(dir, "o.gif"), lgtm, 320); err == nil {
		t.Error("missing input: want error")
	}
	if err := ProcessAnimatedGIF(notGIF, filepath.Join(dir, "o.gif"), lgtm, 320); err == nil {
		t.Error("PNG with .gif extension: want error")
	}
}
