package main

import (
	"fmt"
	"os"

	"github.com/stackquest-hq/godyf/godyf"
	"github.com/stackquest-hq/godyf/pdf"
)

func main() {
	document := pdf.NewPDF()

	// Add metadata to the PDF document
	document.Info.Values["Author"] = godyf.NewString("Jane Doe")
	document.Info.Values["Creator"] = godyf.NewString("pydyf")
	document.Info.Values["Keywords"] = godyf.NewString("some keywords")
	document.Info.Values["Producer"] = godyf.NewString("The producer")
	document.Info.Values["Subject"] = godyf.NewString("An example PDF")
	document.Info.Values["Title"] = godyf.NewString("A PDF containing metadata")

	// Add a page to the document
	page := godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 200, 200),
	})
	document.AddPage(page)

	// Write the document to a PDF file
	file, err := os.Create("metadata.pdf")
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

	fmt.Println("PDF document with metadata written to metadata.pdf")

}
