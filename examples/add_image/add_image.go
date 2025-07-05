package main

import (
	"fmt"
	"os"

	"github.com/stackquest-hq/godyf/godyf"
	"github.com/stackquest-hq/godyf/pdf"
)

func main() {
	document := pdf.NewPDF()
	extra := godyf.NewDictionary(map[string]interface{}{
		"Type":             "/XObject",
		"Subtype":          "/Image",
		"Width":            197,
		"Height":           197,
		"ColorSpace":       "/DeviceRGB",
		"BitsPerComponent": 8,
		"Filter":           "/DCTDecode", // DCTDecode is for JPEG
	})
	image, err := os.ReadFile("examples/add_image/gopher.jpg")
	if err != nil {
		fmt.Printf("Error reading image file: %v\n", err)
		return
	}
	xobject := godyf.NewStream([]interface{}{image}, extra.Values, false)
	document.AddObject(xobject)
	imageStream := godyf.NewStream(nil, nil, false)
	imageStream.PushState()
	imageStream.SetMatrix(100, 0, 0, 100, 100, 100)
	imageStream.DrawXObject("Im1")
	imageStream.PopState()
	document.AddObject(imageStream)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   document.Pages.GetObject().Reference(),
		"MediaBox": godyf.NewArray(0, 0, 595, 842),
		"Resources": godyf.NewDictionary(map[string]interface{}{
			"ProcSet": godyf.NewArray("/PDF", "/ImageB"),
			"XObject": godyf.NewDictionary(map[string]interface{}{
				"Im1": xobject.GetObject().Reference(),
			}),
		}),
		"Contents": imageStream.GetObject().Reference(),
	}))
	file, err := os.Create("document_with_image.pdf")
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
