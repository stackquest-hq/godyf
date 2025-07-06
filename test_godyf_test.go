package godyf_tests

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stackquest-hq/godyf/godyf"
	"github.com/stackquest-hq/godyf/helper"
	"github.com/stackquest-hq/godyf/pdf"
)

func TestFill(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)

	// Create a rectangle path and fill it
	draw.Rectangle(2, 2, 5, 6) // x, y, width, height
	draw.Fill(false)           // Fill with nonzero winding rule

	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))

	// Write PDF to buffer
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__________
__________`)
}

func TestStroke(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(2, 2, 5, 6)
	draw.SetLineWidth(2)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
_KKKKKKK__
_KKKKKKK__
_KK___KK__
_KK___KK__
_KK___KK__
_KK___KK__
_KKKKKKK__
_KKKKKKK__
__________`)

}

func TestLineTo(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.MoveTo(2, 2) // Move to start point
	draw.SetLineWidth(2)
	draw.LineTo(2, 5)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__________
__________
__________
_KK_______
_KK_______
_KK_______
__________
__________
`)
}

func TestSetColorRGBStroke(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(2, 2, 5, 6)
	draw.SetLineWidth(2)
	draw.SetColorRGB(0, 0, 255, true) // Set stroke color to blue
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
_BBBBBBB__
_BBBBBBB__
_BB___BB__
_BB___BB__
_BB___BB__
_BB___BB__
_BBBBBBB__
_BBBBBBB__
__________
`)
}

func TestSetColorRGBFill(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(2, 2, 5, 6)
	draw.SetColorRGB(255, 0, 0, false)
	draw.Fill(false) // Fill with nonzero winding rule
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__RRRRR___
__RRRRR___
__RRRRR___
__RRRRR___
__RRRRR___
__RRRRR___
__________
__________`)
}

func TestSetDash(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.MoveTo(2, 2)
	draw.SetLineWidth(2)
	draw.LineTo(2, 6)
	draw.SetDash([]float64{2.0, 1.0}, 0)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__________
__________
_KK_______
__________
_KK_______
_KK_______
__________
__________`)
}

func TestCurveTo(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.MoveTo(2, 5)
	draw.SetLineWidth(2)
	draw.CurveTo(2, 5, 3, 5, 5, 5) // Example control points
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__________
__________
__KKK_____
__KKK_____
__________
__________
__________
__________`)
}

func TestCurveStartTo(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.MoveTo(2, 5)
	draw.SetLineWidth(2)
	draw.CurveStartTo(3, 5, 5, 5)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__________
__________
__KKK_____
__KKK_____
__________
__________
__________
__________`)
}

func TestCurveEndTo(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.MoveTo(2, 5)
	draw.SetLineWidth(2)
	draw.CurveEndTo(3, 5, 5, 5)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__________
__________
__KKK_____
__KKK_____
__________
__________
__________
__________`)
}

func TestSetMatrix(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.SetMatrix(1, 0, 0, 1, 1, 1)
	draw.MoveTo(2, 2)
	draw.SetLineWidth(2)
	draw.LineTo(2, 5)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__________
__________
__KK______
__KK______
__KK______
__________
__________
__________`)
}

func TestSetState(t *testing.T) {
	document := pdf.NewPDF()

	graphic_state := godyf.NewDictionary(map[string]interface{}{
		"Type": "/ExtGState",
		"LW":   2, // Line width
	})
	document.AddObject(graphic_state)
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(2, 2, 5, 6)
	draw.SetState("GS")
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
		"Resources": godyf.NewDictionary(map[string]interface{}{
			"ExtGState": godyf.NewDictionary(map[string]interface{}{
				"GS": graphic_state.Reference(),
			}),
		}),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
_KKKKKKK__
_KKKKKKK__
_KK___KK__
_KK___KK__
_KK___KK__
_KK___KK__
_KKKKKKK__
_KKKKKKK__
__________`)
}

func TestFillAndStroke(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(2, 2, 5, 6)
	draw.SetLineWidth(2)
	draw.SetColorRGB(0, 0, 255, true)
	draw.FillAndStroke(false) // Fill with nonzero winding rule and stroke
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
_BBBBBBB__
_BBBBBBB__
_BBKKKBB__
_BBKKKBB__
_BBKKKBB__
_BBKKKBB__
_BBBBBBB__
_BBBBBBB__
__________
`)
}

func TestClip(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(3, 3, 5, 6)
	draw.Rectangle(4, 3, 2, 6)
	draw.Clip(false) // Clip with nonzero winding rule
	draw.End()
	draw.MoveTo(0, 5)
	draw.LineTo(10, 5) // Draw a line across the clipped area
	draw.SetLineWidth(2)
	draw.SetColorRGB(255, 0, 0, true)
	draw.SetLineWidth(2)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__________
__________
___RRRRR__
___RRRRR__
__________
__________
__________
__________`)
}

func TestClipEvenOdd(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(3, 3, 5, 6)
	draw.Rectangle(4, 3, 2, 6)
	draw.Clip(true) // Clip with even-odd rule
	draw.End()
	draw.MoveTo(0, 5)
	draw.LineTo(10, 5) // Draw a line across the clipped area
	draw.SetLineWidth(2)
	draw.SetColorRGB(255, 0, 0, true)
	draw.SetLineWidth(2)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__________
__________
___R__RR__
___R__RR__
__________
__________
__________
__________`)
}

func TestClose(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.MoveTo(2, 2)
	draw.LineTo(2, 8)
	draw.LineTo(7, 8)
	draw.LineTo(7, 2)
	draw.Close() // Close the path
	draw.SetColorRGB(0, 0, 255, true)
	draw.SetLineWidth(2)
	draw.Stroke()
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
_BBBBBBB__
_BBBBBBB__
_BB___BB__
_BB___BB__
_BB___BB__
_BB___BB__
_BBBBBBB__
_BBBBBBB__
__________`)
}

func TestStrokeAndClose(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.MoveTo(2, 2)
	draw.LineTo(2, 8)
	draw.LineTo(7, 8)
	draw.LineTo(7, 2)
	draw.SetColorRGB(0, 0, 255, true)
	draw.SetLineWidth(2)
	draw.StrokeAndClose() // Stroke and close the path
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
_BBBBBBB__
_BBBBBBB__
_BB___BB__
_BB___BB__
_BB___BB__
_BB___BB__
_BBBBBBB__
_BBBBBBB__
__________`)
}

func TestFillStrokeAndClose(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.MoveTo(2, 2)
	draw.LineTo(2, 8)
	draw.LineTo(7, 8)
	draw.LineTo(7, 2)
	draw.SetColorRGB(255, 0, 0, false)
	draw.SetColorRGB(0, 0, 255, true)
	draw.SetLineWidth(2)
	draw.FillStrokeAndClose(false) // Fill and stroke with nonzero winding rule
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
_BBBBBBB__
_BBBBBBB__
_BBRRRBB__
_BBRRRBB__
_BBRRRBB__
_BBRRRBB__
_BBBBBBB__
_BBBBBBB__
__________`)
}

func TestPushPopState(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(2, 2, 5, 6)
	draw.PushState()
	draw.Rectangle(4, 4, 2, 2)
	draw.SetColorRGB(255, 0, 0, false)
	draw.PopState()
	draw.Fill(false)
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__________
__________`)
}

func TestTypes(t *testing.T) {
	document := pdf.NewPDF()
	draw := godyf.NewStream(nil, nil, false)
	draw.Rectangle(2, 2.0, 5, 6)
	draw.SetLineWidth(2.3456)
	draw.Fill(false)
	document.AddObject(draw)
	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}
	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__________
__________`)
}

func TestCompress(t *testing.T) {
	document := pdf.NewPDF()

	// Test uncompressed stream - should contain raw data
	draw := godyf.NewStream(nil, nil, false) // compress=false
	draw.Rectangle(2, 2, 5, 6)
	draw.Fill(false)
	drawData := draw.Data()
	if !bytes.Contains(drawData, []byte("2 2 5 6")) {
		t.Fatalf("Expected uncompressed stream to contain '2 2 5 6', but it didn't")
	}

	// Test compressed stream - should NOT contain raw data
	drawCompressed := godyf.NewStream(nil, nil, true) // compress=true
	drawCompressed.Rectangle(2, 2, 5, 6)
	drawCompressed.Fill(false)
	drawCompressedData := drawCompressed.Data()
	if bytes.Contains(drawCompressedData, []byte("2 2 5 6")) {
		t.Fatalf("Expected compressed stream to NOT contain '2 2 5 6', but it did")
	}
	document.AddObject(drawCompressed)

	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(drawCompressed.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
	}))

	// Write PDF to buffer
	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	helper.AssertPixelsT(t, buf.Bytes(), `
__________
__________
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__KKKKK___
__________
__________`)
}

func TestText(t *testing.T) {
	document := pdf.NewPDF()
	font := godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Font",
		"Subtype":  "/Type1",
		"Name":     "/F1",
		"BaseFont": "/Helvetica",
		"Encoding": "/MacRomanEncoding",
	})
	document.AddObject(font)

	draw := godyf.NewStream(nil, nil, false)
	draw.BeginText()
	draw.SetFontSize("F1", 200)
	draw.SetTextMatrix(1, 0, 0, 1, -20, 5)
	draw.ShowTextString("l")
	draw.ShowTextString("É")

	draw.EndText()

	document.AddObject(draw)

	document.AddPage(godyf.NewDictionary(map[string]interface{}{
		"Type":     "/Page",
		"Parent":   string(document.Pages.Reference()),
		"Contents": string(draw.Reference()),
		"MediaBox": godyf.NewArray(0, 0, 10, 10),
		"Resources": godyf.NewDictionary(map[string]interface{}{
			"ProcSet": godyf.NewArray("/PDF", "/Text"),
			"Font": godyf.NewDictionary(map[string]interface{}{
				"F1": font.Reference(),
			}),
		}),
	}))

	var buf bytes.Buffer
	err := document.Write(&buf, nil, nil, false)
	if err != nil {
		t.Fatalf("Failed to write PDF: %v", err)
	}

	helper.AssertPixelsT(t, buf.Bytes(), `
KKKKKKKKKK
KKKKKKKKKK
KKKKKKKKKK
KKKKKKKKKK
KKKKKKKKKK
zzzzzzzzzz
__________
__________
__________
__________`)
}

func TestNoIdentifier(t *testing.T) {
	document := pdf.NewPDF()
	var pdf bytes.Buffer
	document.Write(&pdf, nil, false, false)
	re := regexp.MustCompile(`/ID \[\(([0-9a-f]{32})\) \(([0-9a-f]{32})\)\]`)

	if re.Match(pdf.Bytes()) {
		t.Fatal("Unexpected /ID found in PDF when identifier=False")
	}
}

func TestWithIdentifier(t *testing.T) {
	document := pdf.NewPDF()
	var pdf bytes.Buffer
	document.Write(&pdf, nil, true, false)
	re := regexp.MustCompile(`/ID \[\(([0-9a-f]{32})\) \(([0-9a-f]{32})\)\]`)

	if !re.Match(pdf.Bytes()) {
		t.Fatal("Expected /ID not found in PDF when identifier=True")
	}

	matches := re.FindSubmatch(pdf.Bytes())
	if len(matches) != 3 {
		t.Fatal("Expected two matches for /ID")
	}
}

func TestCustomIdentifier(t *testing.T) {
	document := pdf.NewPDF()
	var pdf bytes.Buffer
	var identifier []byte = []byte("abc")
	document.Write(&pdf, nil, identifier, false)
	re := regexp.MustCompile(`/ID \[\(abc\) \(([0-9a-f]{32})\)\]`)

	if !re.Match(pdf.Bytes()) {
		t.Fatal("Expected custom /ID not found in PDF")
	}

	matches := re.FindSubmatch(pdf.Bytes())
	if len(matches) != 2 {
		t.Fatalf("Expected 2 matches (full match + one capture group) for custom /ID, got %d", len(matches))
	}
}

func TestVersion(t *testing.T) {
	document := pdf.NewPDF()
	var pdf bytes.Buffer
	document.Write(&pdf, []byte("1.7"), nil, false)

	re := regexp.MustCompile(`^%PDF-1\.7\r?\n`)
	if !re.Match(pdf.Bytes()) {
		t.Fatal("Expected PDF version 1.7 not found in PDF")
	}
}
