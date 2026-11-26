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

func main() {
	var (
		lgtmPath string
		width    int
	)

	// Parse flags
	nameFlag := flag.String("input", "", "Input file name in input/ (extension optional)")
	targetWidthFlag := flag.Int("width", 320, "Target width for resizing")
	flag.Parse()

	if *nameFlag == "" {
		log.Fatal("You must specify --input")
	}

	inputPath, err := resolveInput("input", *nameFlag)
	if err != nil {
		log.Fatal(err)
	}
	ext := strings.ToLower(filepath.Ext(inputPath))
	stem := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))

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

	switch ext {
	case ".gif":
		outputPath := fmt.Sprintf("output/%d_lgtm_%s.gif", width, stem)
		if err := ProcessAnimatedGIF(inputPath, outputPath, lgtmPath, width); err != nil {
			log.Fatalf("GIF processing failed: %v", err)
		}
		fmt.Printf("Animated GIF generated successfully: %s\n", outputPath)
	default:
		outputPath := fmt.Sprintf("output/%d_lgtm_%s.webp", width, stem)
		if err := ProcessStaticImage(inputPath, outputPath, lgtmPath, width); err != nil {
			log.Fatalf("Static processing failed: %v", err)
		}
		fmt.Printf("WebP image generated successfully: %s\n", outputPath)
	}
}

// resolveInput finds the file in dir for name. name may omit the extension;
// if several files share the same name, it returns an error listing them.
func resolveInput(dir, name string) (string, error) {
	exact := filepath.Join(dir, name)
	if info, err := os.Stat(exact); err == nil && !info.IsDir() {
		return exact, nil
	}

	matches, err := filepath.Glob(filepath.Join(dir, globEscape(name)+".*"))
	if err != nil {
		return "", fmt.Errorf("failed to search input: %w", err)
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

func globEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `*`, `\*`, `?`, `\?`, `[`, `\[`).Replace(s)
}
