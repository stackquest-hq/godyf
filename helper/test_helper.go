package helper

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// PIXELS_BY_CHAR maps ASCII chars to expected RGB colors
var PIXELS_BY_CHAR = map[rune]*color.RGBA{
	'_': {R: 255, G: 255, B: 255, A: 255}, // white
	'R': {R: 255, G: 0, B: 0, A: 255},     // red
	'B': {R: 0, G: 0, B: 255, A: 255},     // blue
	'G': {R: 0, G: 255, B: 0, A: 255},     // green
	'K': {R: 0, G: 0, B: 0, A: 255},       // black
	'z': nil,                              // wildcard (any color)
}

// AssertPixels compares a PDF rendering with a reference ASCII art pattern
func AssertPixels(pdfData []byte, referencePixels string) {
	AssertPixelsT(nil, pdfData, referencePixels)
}

// AssertPixelsT compares a PDF rendering with a reference ASCII art pattern (for testing)
func AssertPixelsT(t *testing.T, pdfData []byte, referencePixels string) {
	// Run Ghostscript to render PDF to PNG
	cmd := exec.Command("gs", "-q", "-dNOPAUSE", "-dSAFER", "-sDEVICE=png16m",
		"-r576", "-dDownScaleFactor=8", "-sOutputFile=-", "-")

	cmd.Stdin = bytes.NewReader(pdfData)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		if t != nil {
			t.Fatalf("failed to run Ghostscript: %v", err)
		} else {
			log.Fatalf("failed to run Ghostscript: %v", err)
		}
	}

	// Decode the PNG image
	img, _, err := image.Decode(bytes.NewReader(out.Bytes()))
	if err != nil {
		if t != nil {
			t.Fatalf("failed to decode PNG: %v", err)
		} else {
			log.Fatalf("failed to decode PNG: %v", err)
		}
	}

	// Parse the reference pixels into lines
	lines := []string{}
	for _, line := range strings.Split(referencePixels, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}

	width := len(lines[0])
	height := len(lines)
	for _, line := range lines {
		if len(line) != width {
			if t != nil {
				t.Fatalf("reference lines are not the same length")
			} else {
				log.Fatalf("reference lines are not the same length")
			}
		}
	}

	if img.Bounds().Dx() != width || img.Bounds().Dy() != height {
		if t != nil {
			t.Fatalf("reference is %dx%d, image is %dx%d",
				width, height, img.Bounds().Dx(), img.Bounds().Dy())
		} else {
			log.Fatalf("reference is %dx%d, image is %dx%d",
				width, height, img.Bounds().Dx(), img.Bounds().Dy())
		}
	}

	// Compare pixel colors
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := rune(lines[y][x])
			refColor := PIXELS_BY_CHAR[c]
			gotColor := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)

			if refColor == nil {
				continue
			}
			if gotColor.R != refColor.R || gotColor.G != refColor.G || gotColor.B != refColor.B {
				// Save actual and reference images for debugging
				testName := os.Getenv("TEST_NAME")
				if testName == "" {
					testName = "test"
				}
				WritePNG(testName+"-actual", img)
				refImg := BuildReferenceImage(lines, width, height)
				WritePNG(testName+"-reference", refImg)
				if t != nil {
					t.Fatalf("Pixel mismatch at (%d,%d): expected %v, got %v",
						x, y, refColor, gotColor)
				} else {
					log.Fatalf("Pixel mismatch at (%d,%d): expected %v, got %v",
						x, y, refColor, gotColor)
				}
			}
		}
	}
}

// buildReferenceImage builds an image.Image from ASCII reference
func BuildReferenceImage(lines []string, width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y, line := range lines {
		for x, c := range line {
			refColor := PIXELS_BY_CHAR[c]
			if refColor == nil {
				img.Set(x, y, color.RGBA{255, 255, 255, 255})
			} else {
				img.Set(x, y, *refColor)
			}
		}
	}
	return img
}

// writePNG writes the image to results/{name}.png
func WritePNG(name string, img image.Image) {
	dir := filepath.Join(filepath.Dir(os.Args[0]), "results")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create results dir: %v", err)
	}
	filePath := filepath.Join(dir, name+".png")
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatalf("failed to create file: %v", err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		log.Fatalf("failed to encode PNG: %v", err)
	}
	log.Printf("Wrote image: %s", filePath)
}
