package main

import (
	"fmt"
	"os"

	"github.com/stackquest-hq/godyf/godyf"
	"github.com/stackquest-hq/godyf/pdf"
)

func main() {
	document := pdf.NewPDF()
	font := godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Font",
		"Subtype":  "/Type1",
		"Name":     "/F1",
		"BaseFont": "/Helvetica",
		"Encoding": "/MacRomanEncoding",
	})
	document.AddObject(font)
	text := godyf.NewStream(nil, nil, false)
	text.BeginText()
	text.SetFontSize("F1", 20)
	text.SetTextMatrix(1, 0, 0, 1, 10, 90)
	text.ShowTextString("Bœuf grillé & café")
	text.EndText()
	document.AddObject(text)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 595, 842),
		"Contents": text.Reference(),
		"Resources": godyf.NewDictionary(map[string]interface{}{
			"ProcSet": godyf.NewArray("/PDF", "/Text"),
			"Font": godyf.NewDictionary(map[string]interface{}{
				"F1": font.Reference(),
			}),
		}),
	}))
	file, err := os.Create("document3.pdf")
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
