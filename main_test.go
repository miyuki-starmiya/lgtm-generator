package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsPattern(t *testing.T) {
	tests := map[string]bool{
		"aoba01":      false,
		"aoba01.jpg":  false,
		"rose0*":      true,
		"rose0?.jpg":  true,
		"rose0[1-2]":  true,
		"with-dash_1": false,
	}
	for in, want := range tests {
		if got := isPattern(in); got != want {
			t.Errorf("isPattern(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestOutputPathFor(t *testing.T) {
	tests := []struct {
		in    string
		width int
		want  string
	}{
		{"input/aoba01.jpg", 320, "output/320_lgtm_aoba01.webp"},
		{"input/coco01.jpeg", 500, "output/500_lgtm_coco01.webp"},
		{"input/chisato01.webp", 320, "output/320_lgtm_chisato01.webp"},
		{"input/kurumi00.gif", 320, "output/320_lgtm_kurumi00.gif"},
		{"input/KURUMI.GIF", 320, "output/320_lgtm_KURUMI.gif"},
		{"input/a.b.png", 320, "output/320_lgtm_a.b.webp"},
	}
	for _, tt := range tests {
		if got := outputPathFor(tt.in, tt.width); got != tt.want {
			t.Errorf("outputPathFor(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
		}
	}
}

func TestResolveInput(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"aoba01.jpg", "chisato01.jpg", "chisato01.webp", "a[1].png", "foo.png"} {
		touch(t, filepath.Join(dir, f))
	}
	// A directory that matches "foo.*" must not count as a candidate.
	if err := os.Mkdir(filepath.Join(dir, "foo.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		want    string
		wantErr string
	}{
		{name: "aoba01", want: "aoba01.jpg"},
		{name: "aoba01.jpg", want: "aoba01.jpg"},
		{name: "chisato01.webp", want: "chisato01.webp"},
		{name: "a[1]", want: "a[1].png"},
		{name: "foo", want: "foo.png"},
		{name: "chisato01", wantErr: "chisato01.jpg, chisato01.webp"},
		{name: "nothing", wantErr: "input not found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveInput(dir, tt.name)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != filepath.Join(dir, tt.want) {
				t.Errorf("got %q, want %q", got, filepath.Join(dir, tt.want))
			}
		})
	}
}

func TestMatchInputs(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"rose01.webp", "rose02.jpeg", "rose03.jpg", "lily01.png"} {
		touch(t, filepath.Join(dir, f))
	}
	if err := os.Mkdir(filepath.Join(dir, "rose04.d"), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		pattern string
		want    []string
		wantErr string
	}{
		{pattern: "rose0*.*", want: []string{"rose01.webp", "rose02.jpeg", "rose03.jpg"}},
		{pattern: "rose0*", want: []string{"rose01.webp", "rose02.jpeg", "rose03.jpg"}},
		// "rose0?" alone matches nothing, so it falls back to "rose0?.*".
		{pattern: "rose0?", want: []string{"rose01.webp", "rose02.jpeg", "rose03.jpg"}},
		{pattern: "rose0[12]", want: []string{"rose01.webp", "rose02.jpeg"}},
		{pattern: "*.png", want: []string{"lily01.png"}},
		{pattern: "zzz*", wantErr: "no input matches"},
		{pattern: "[", wantErr: "invalid pattern"},
	}
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got, err := matchInputs(dir, tt.pattern)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var names []string
			for _, m := range got {
				names = append(names, filepath.Base(m))
			}
			if strings.Join(names, ",") != strings.Join(tt.want, ",") {
				t.Errorf("got %v, want %v", names, tt.want)
			}
		})
	}
}

func runCLI(args ...string) (code int, stdout, stderr string) {
	var out, errOut bytes.Buffer
	code = run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestRunHelp(t *testing.T) {
	for _, arg := range []string{"--help", "-h"} {
		t.Run(arg, func(t *testing.T) {
			code, stdout, stderr := runCLI(arg)
			if code != 0 {
				t.Errorf("exit code = %d, want 0", code)
			}
			if stderr != "" {
				t.Errorf("stderr = %q, want empty", stderr)
			}
			for _, want := range []string{"Usage:", "Examples:", "--input 'rose0*'", "-input", "-width", ".webp", ".gif"} {
				if !strings.Contains(stdout, want) {
					t.Errorf("help does not contain %q:\n%s", want, stdout)
				}
			}
			if n := strings.Count(stdout, "Usage:"); n != 1 {
				t.Errorf("help printed %d times, want once", n)
			}
		})
	}
}

func TestRunUsageErrors(t *testing.T) {
	setupWorkspace(t)
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"no input", nil, "you must specify --input"},
		{"unknown flag", []string{"--bogus"}, "Usage:"},
		{"width without asset", []string{"--input", "a", "--width", "400"}, "assets/400_lgtm.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, stderr := runCLI(tt.args...)
			if code != 2 {
				t.Errorf("exit code = %d, want 2", code)
			}
			if !strings.Contains(stderr, tt.wantErr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr, tt.wantErr)
			}
		})
	}
}

func TestRunSingle(t *testing.T) {
	dir := setupWorkspace(t)
	writePNG(t, filepath.Join(dir, "input", "aoba01.png"), 640, 480, srcColor)
	writeGIF(t, filepath.Join(dir, "input", "kurumi00.gif"), 640, 480, 2)

	tests := []struct {
		input string
		want  string
	}{
		{"aoba01", "output/320_lgtm_aoba01.webp"},
		{"aoba01.png", "output/320_lgtm_aoba01.webp"},
		{"kurumi00", "output/320_lgtm_kurumi00.gif"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			code, stdout, stderr := runCLI("--input", tt.input)
			if code != 0 {
				t.Fatalf("exit code = %d, stderr = %q", code, stderr)
			}
			if !strings.Contains(stdout, "Generated: "+tt.want) {
				t.Errorf("stdout = %q", stdout)
			}
			if _, err := os.Stat(tt.want); err != nil {
				t.Error(err)
			}
		})
	}

	code, _, stderr := runCLI("--input", "nothing")
	if code != 1 || !strings.Contains(stderr, "input not found") {
		t.Errorf("missing input: code = %d, stderr = %q", code, stderr)
	}
}

func TestRunBatch(t *testing.T) {
	dir := setupWorkspace(t)
	for _, f := range []string{"rose01.png", "rose02.png", "rose03.png"} {
		writePNG(t, filepath.Join(dir, "input", f), 640, 480, srcColor)
	}

	code, stdout, stderr := runCLI("--input", "rose0*")
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, "Done: 3 succeeded, 0 failed") {
		t.Errorf("stdout = %q", stdout)
	}
	for _, f := range []string{"rose01", "rose02", "rose03"} {
		if _, err := os.Stat("output/320_lgtm_" + f + ".webp"); err != nil {
			t.Error(err)
		}
	}
}

func TestRunBatchContinuesAfterFailure(t *testing.T) {
	dir := setupWorkspace(t)
	writePNG(t, filepath.Join(dir, "input", "chisato01.png"), 640, 480, srcColor)
	// Same name as chisato01.png, so it would overwrite the same output.
	writePNG(t, filepath.Join(dir, "input", "chisato01.webp"), 640, 480, srcColor)
	// Not an image, so decoding fails.
	if err := os.WriteFile(filepath.Join(dir, "input", "chisato02.jpg"), []byte("not an image"), 0o644); err != nil {
		t.Fatal(err)
	}
	writePNG(t, filepath.Join(dir, "input", "chisato03.png"), 640, 480, srcColor)

	code, stdout, _ := runCLI("--input", "chisato*")
	if code != 1 {
		t.Errorf("exit code = %d, want 1", code)
	}
	for _, want := range []string{
		"Generated: output/320_lgtm_chisato01.webp",
		"Skipped: input/chisato01.webp",
		"Failed: input/chisato02.jpg",
		"Generated: output/320_lgtm_chisato03.webp",
		"Done: 2 succeeded, 2 failed",
	} {
		if !strings.Contains(stdout, want) {
			t.Errorf("stdout does not contain %q:\n%s", want, stdout)
		}
	}
}
