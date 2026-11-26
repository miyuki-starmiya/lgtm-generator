// main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/webp"
)

const inputDir = "input"

func main() {
	var (
		lgtmPath string
		width    int
	)

	// Parse flags
	nameFlag := flag.String("input", "", "Input file name in input/ (extension optional), or a glob pattern such as 'rose0*'")
	targetWidthFlag := flag.Int("width", 320, "Target width for resizing")
	flag.Parse()

	if *nameFlag == "" {
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

	if !isPattern(*nameFlag) {
		inputPath, err := resolveInput(inputDir, *nameFlag)
		if err != nil {
			log.Fatal(err)
		}
		outputPath, err := process(inputPath, lgtmPath, width)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Generated: %s\n", outputPath)
		return
	}

	inputPaths, err := matchInputs(inputDir, *nameFlag)
	if err != nil {
		log.Fatal(err)
	}

	// Inputs that differ only by extension would write the same output file.
	written := map[string]string{}
	failed := 0
	for _, inputPath := range inputPaths {
		out := outputPathFor(inputPath, width)
		if prev, ok := written[out]; ok {
			fmt.Printf("Skipped: %s (same output as %s: %s)\n", inputPath, prev, out)
			failed++
			continue
		}
		written[out] = inputPath

		outputPath, err := process(inputPath, lgtmPath, width)
		if err != nil {
			fmt.Printf("Failed: %s: %v\n", inputPath, err)
			failed++
			continue
		}
		fmt.Printf("Generated: %s\n", outputPath)
	}
	fmt.Printf("Done: %d succeeded, %d failed\n", len(inputPaths)-failed, failed)
	if failed > 0 {
		os.Exit(1)
	}
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
