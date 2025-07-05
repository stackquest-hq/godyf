package main

import (
	"fmt"
	"os"

	"github.com/stackquest-hq/godyf/godyf"
	"github.com/stackquest-hq/godyf/pdf"
)

func main() {
	document := pdf.NewPDF()

	draw := godyf.NewStream(nil, nil, false)
	draw.SetColorRGB(1.0, 0.0, 0.0, false) // Set nonstroking color to red
	draw.SetColorRGB(0.0, 1.0, 0.0, true)  // Set stroking color to green
	draw.Rectangle(100, 100, 50, 70)
	draw.SetDash([]float64{2, 1}, 0) // Set dash pattern: 2 points on, 1 point off
	draw.Stroke()
	draw.Rectangle(50, 50, 20, 40)
	draw.SetDash([]float64{}, 0) // Reset to solid line
	draw.SetLineWidth(10)
	draw.SetMatrix(1, 0, 0, 1, 80, 80) // Move to (80, 80)
	draw.Fill(false)                   // Fill the rectangle without stroke
	document.AddObject(draw)
	page := godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   document.Pages.Reference(),
		"Contents": draw.Reference(),
		"MediaBox": godyf.NewArray(0, 0, 200, 200),
	})
	document.AddPage(page)
	file, err := os.Create("document_with_color.pdf")
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()
	err = document.Write(file, nil, nil, false)
	if err != nil {
		fmt.Printf("Error writing PDF: %v\n", err)
		return
	}
}
