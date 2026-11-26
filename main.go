// main.go
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"
)

const inputDir = "input"

const usageText = `lgtm-generator overlays an LGTM image on pictures in input/ and writes them to output/.

Usage:
  go run . --input <name> [--width <320|500>]

Input:
  <name> is a file name in input/. The extension can be omitted.
  If several files share the name (e.g. a.jpg and a.webp), specify the extension.
  A glob pattern (*, ?, [...]) processes every matching file. Quote it so the
  shell does not expand it in the current directory.

Output:
  Static images (jpg, png, webp, ...) -> output/<width>_lgtm_<name>.webp (lossy, quality 80)
  GIF                                 -> output/<width>_lgtm_<name>.gif  (animated)

Examples:
  go run . --input aoba01
  go run . --input coco01.jpeg --width 500
  go run . --input 'rose0*'

Flags:
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the CLI and returns the exit code: 0 on success, 1 when
// processing fails, 2 on invalid usage.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("lgtm-generator", flag.ContinueOnError)
	fs.SetOutput(stderr)
	name := fs.String("input", "", "file name in input/ (extension optional), or a quoted glob pattern such as 'rose0*'")
	width := fs.Int("width", 320, "target width in px; assets/<width>_lgtm.png must exist (320, 500)")
	// Parse calls fs.Usage itself on -h and on bad flags; print usage below
	// instead, so --help goes to stdout.
	fs.Usage = func() {}
	usage := func(w io.Writer) {
		fmt.Fprint(w, usageText)
		fs.SetOutput(w)
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			usage(stdout)
			return 0
		}
		usage(stderr)
		return 2
	}

	if *name == "" {
		fmt.Fprintln(stderr, "error: you must specify --input (see --help)")
		return 2
	}

	lgtmPath := fmt.Sprintf("assets/%d_lgtm.png", *width)
	if _, err := os.Stat(lgtmPath); err != nil {
		fmt.Fprintf(stderr, "error: LGTM image does not exist: %s\n", lgtmPath)
		return 2
	}

	if !isPattern(*name) {
		inputPath, err := resolveInput(inputDir, *name)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		outputPath, err := process(inputPath, lgtmPath, *width)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Generated: %s\n", outputPath)
		return 0
	}

	inputPaths, err := matchInputs(inputDir, *name)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	// Inputs that differ only by extension would write the same output file.
	written := map[string]string{}
	failed := 0
	for _, inputPath := range inputPaths {
		out := outputPathFor(inputPath, *width)
		if prev, ok := written[out]; ok {
			fmt.Fprintf(stdout, "Skipped: %s (same output as %s: %s)\n", inputPath, prev, out)
			failed++
			continue
		}
		written[out] = inputPath

		outputPath, err := process(inputPath, lgtmPath, *width)
		if err != nil {
			fmt.Fprintf(stdout, "Failed: %s: %v\n", inputPath, err)
			failed++
			continue
		}
		fmt.Fprintf(stdout, "Generated: %s\n", outputPath)
	}
	fmt.Fprintf(stdout, "Done: %d succeeded, %d failed\n", len(inputPaths)-failed, failed)
	if failed > 0 {
		return 1
	}
	return 0
}

// process writes the LGTM image for inputPath and returns the output path.
// GIFs stay animated GIFs; every other image becomes WebP.
func process(inputPath, lgtmPath string, width int) (string, error) {
	outputPath := outputPathFor(inputPath, width)
	if isGIF(inputPath) {
		if err := ProcessAnimatedGIF(inputPath, outputPath, lgtmPath, width); err != nil {
			return "", fmt.Errorf("GIF processing failed: %w", err)
		}
		return outputPath, nil
	}
	if err := ProcessStaticImage(inputPath, outputPath, lgtmPath, width); err != nil {
		return "", fmt.Errorf("static processing failed: %w", err)
	}
	return outputPath, nil
}

func outputPathFor(inputPath string, width int) string {
	stem := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	ext := ".webp"
	if isGIF(inputPath) {
		ext = ".gif"
	}
	return fmt.Sprintf("output/%d_lgtm_%s%s", width, stem, ext)
}

func isGIF(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".gif"
}

func isPattern(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

// resolveInput finds the file in dir for name. name may omit the extension;
// if several files share the same name, it returns an error listing them.
func resolveInput(dir, name string) (string, error) {
	exact := filepath.Join(dir, name)
	if info, err := os.Stat(exact); err == nil && !info.IsDir() {
		return exact, nil
	}

	matches, err := globFiles(filepath.Join(dir, globEscape(name)+".*"))
	if err != nil {
		return "", err
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("input not found: %s", exact)
	case 1:
		return matches[0], nil
	default:
		names := make([]string, len(matches))
		for i, m := range matches {
			names[i] = filepath.Base(m)
		}
		return "", fmt.Errorf("multiple inputs match %q: %s (specify the extension)", name, strings.Join(names, ", "))
	}
}

// matchInputs returns the files in dir that match pattern. A pattern without
// an extension, such as "rose0*", also matches "rose0*.*".
func matchInputs(dir, pattern string) ([]string, error) {
	matches, err := globFiles(filepath.Join(dir, pattern))
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 && filepath.Ext(pattern) == "" {
		if matches, err = globFiles(filepath.Join(dir, pattern+".*")); err != nil {
			return nil, err
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no input matches: %s", filepath.Join(dir, pattern))
	}
	return matches, nil
}

// globFiles is filepath.Glob without directories.
func globFiles(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	files := matches[:0]
	for _, m := range matches {
		if info, err := os.Stat(m); err == nil && !info.IsDir() {
			files = append(files, m)
		}
	}
	return files, nil
}

func globEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `*`, `\*`, `?`, `\?`, `[`, `\[`).Replace(s)
}
