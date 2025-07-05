package main

import (
	"fmt"
	"os"

	"github.com/stackquest-hq/godyf/godyf"
	"github.com/stackquest-hq/godyf/pdf"
)

func main() {
	document := pdf.NewPDF()

	page := godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 595, 842),
	})

	document.AddPage(page)

	// Write the document to a PDF file
	file, err := os.Create("document.pdf")
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

	fmt.Println("PDF document written to document.pdf")
}
