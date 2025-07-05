package pdf

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/stackquest-hq/godyf/godyf"
)

// ObjectWrapper wraps a basic Object to implement PDFObject interface
type ObjectWrapper struct {
	*godyf.Object
}

// Data returns empty data for basic object wrapper
func (o *ObjectWrapper) Data() []byte {
	return []byte{}
}

// GetObject returns the underlying Object struct
func (o *ObjectWrapper) GetObject() *godyf.Object {
	return o.Object
}

// SetObject sets the underlying Object struct
func (o *ObjectWrapper) SetObject(obj *godyf.Object) {
	o.Object = obj
}

// PDF represents a PDF document
type PDF struct {
	// Objects contains all PDF objects
	Objects []godyf.PDFObject
	// Pages dictionary containing the PDF's pages
	Pages *godyf.Dictionary
	// Info dictionary containing the PDF's metadata
	Info *godyf.Dictionary
	// Catalog dictionary containing references to other objects
	Catalog *godyf.Dictionary
	// Current position in the PDF
	CurrentPosition int
	// Position of the cross reference table
	XRefPosition int
}

// NewPDF creates a new PDF document
func NewPDF() *PDF {
	pdf := &PDF{
		Objects:         make([]godyf.PDFObject, 0),
		CurrentPosition: 0,
	}

	// Create zero object (first object is always a free object)
	zeroObject := godyf.NewObject()
	zeroObject.Generation = 65535
	zeroObject.Free = 'f'
	pdf.AddObject(&ObjectWrapper{zeroObject})

	// Create Pages dictionary
	pdf.Pages = godyf.NewDictionary(map[string]interface{}{
		"Type":  "/Pages",
		"Kids":  godyf.NewArray(),
		"Count": 0,
	})
	pdf.AddObject(pdf.Pages)

	// Create Info dictionary
	pdf.Info = godyf.NewDictionary(map[string]interface{}{})
	pdf.AddObject(pdf.Info)

	// Create Catalog dictionary
	pdf.Catalog = godyf.NewDictionary(map[string]interface{}{
		"Type":  "/Catalog",
		"Pages": pdf.Pages.GetObject().Reference(),
	})
	pdf.AddObject(pdf.Catalog)

	return pdf
}

// AddPage adds a page to the PDF
func (p *PDF) AddPage(page *godyf.Dictionary) {
	// Increment page count
	p.Pages.Values["Count"] = p.Pages.Values["Count"].(int) + 1

	// Add page object
	p.AddObject(page)

	// Add page reference to Kids array
	kids := p.Pages.Values["Kids"].(*godyf.Array)
	kids.Elements = append(kids.Elements, page.GetObject().Number)
	kids.Elements = append(kids.Elements, 0)
	kids.Elements = append(kids.Elements, "R")
}

// AddObject adds an object to the PDF
func (p *PDF) AddObject(obj godyf.PDFObject) {
	objBase := obj.GetObject()
	objBase.Number = len(p.Objects)
	p.Objects = append(p.Objects, obj)
}

// PageReferences returns the page references
func (p *PDF) PageReferences() [][]byte {
	kids := p.Pages.Values["Kids"].(*godyf.Array)
	var references [][]byte

	// Extract every third element (object numbers)
	for i := 0; i < len(kids.Elements); i += 3 {
		if objNum, ok := kids.Elements[i].(int); ok {
			ref := fmt.Sprintf("%d 0 R", objNum)
			references = append(references, []byte(ref))
		}
	}

	return references
}

// WriteLine writes a line to the output and updates current position
func (p *PDF) WriteLine(content []byte, output io.Writer) error {
	p.CurrentPosition += len(content) + 1
	_, err := output.Write(content)
	if err != nil {
		return err
	}
	_, err = output.Write([]byte("\n"))
	return err
}

// Write writes the PDF to the output
func (p *PDF) Write(output io.Writer, version []byte, identifier interface{}, compress bool) error {
	// Convert version to bytes, default to "1.7"
	if version == nil {
		version = []byte("1.7")
	}

	// Write header
	header := append([]byte("%PDF-"), version...)
	if err := p.WriteLine(header, output); err != nil {
		return err
	}

	// Write binary marker
	if err := p.WriteLine([]byte("%\xf0\x9f\x96\xa4"), output); err != nil {
		return err
	}

	if bytes.Compare(version, []byte("1.5")) >= 0 && compress {
		return p.writeCompressed(output, identifier)
	} else {
		return p.writeUncompressed(output, identifier)
	}
}

// writeUncompressed writes PDF without compression
func (p *PDF) writeUncompressed(output io.Writer, identifier interface{}) error {
	// Write all non-free PDF objects
	for _, obj := range p.Objects {
		objBase := obj.GetObject()
		if objBase.Free == 'f' {
			continue
		}
		objBase.Offset = p.CurrentPosition
		indirect := objBase.Indirect(obj.Data())
		if err := p.WriteLine(indirect, output); err != nil {
			return err
		}
	}

	// Write cross-reference table
	p.XRefPosition = p.CurrentPosition
	if err := p.WriteLine([]byte("xref"), output); err != nil {
		return err
	}

	xrefHeader := fmt.Sprintf("0 %d", len(p.Objects))
	if err := p.WriteLine([]byte(xrefHeader), output); err != nil {
		return err
	}

	// Write xref entries
	for _, obj := range p.Objects {
		objBase := obj.GetObject()
		entry := fmt.Sprintf("%010d %05d %c ", objBase.Offset, objBase.Generation, objBase.Free)
		if err := p.WriteLine([]byte(entry), output); err != nil {
			return err
		}
	}

	// Write trailer
	if err := p.WriteLine([]byte("trailer"), output); err != nil {
		return err
	}
	if err := p.WriteLine([]byte("<<"), output); err != nil {
		return err
	}

	sizeEntry := fmt.Sprintf("/Size %d", len(p.Objects))
	if err := p.WriteLine([]byte(sizeEntry), output); err != nil {
		return err
	}

	rootEntry := append([]byte("/Root "), p.Catalog.GetObject().Reference()...)
	if err := p.WriteLine(rootEntry, output); err != nil {
		return err
	}

	infoEntry := append([]byte("/Info "), p.Info.GetObject().Reference()...)
	if err := p.WriteLine(infoEntry, output); err != nil {
		return err
	}

	// Handle identifier if provided
	if identifier != nil {
		if err := p.writeIdentifier(output, identifier); err != nil {
			return err
		}
	}

	if err := p.WriteLine([]byte(">>"), output); err != nil {
		return err
	}

	// Write startxref
	if err := p.WriteLine([]byte("startxref"), output); err != nil {
		return err
	}

	xrefPos := fmt.Sprintf("%d", p.XRefPosition)
	if err := p.WriteLine([]byte(xrefPos), output); err != nil {
		return err
	}

	return p.WriteLine([]byte("%%EOF"), output)
}

// writeCompressed writes PDF with compression (placeholder for now)
func (p *PDF) writeCompressed(output io.Writer, identifier interface{}) error {
	// TODO: Implement compressed writing similar to Python version
	// For now, fall back to uncompressed
	return p.writeUncompressed(output, identifier)
}

// writeIdentifier writes the PDF identifier
func (p *PDF) writeIdentifier(output io.Writer, identifier interface{}) error {
	// Calculate data hash
	var data bytes.Buffer
	for _, obj := range p.Objects {
		objBase := obj.GetObject()
		if objBase.Free != 'f' {
			data.Write(obj.Data())
		}
	}

	hasher := md5.New()
	hasher.Write(data.Bytes())
	dataHash := hex.EncodeToString(hasher.Sum(nil))

	var idBytes []byte
	if identifier == true {
		idBytes = []byte(dataHash)
	} else if idStr, ok := identifier.(string); ok {
		idBytes = []byte(idStr)
	} else if idBytes, ok = identifier.([]byte); !ok {
		return fmt.Errorf("invalid identifier type")
	}

	idString1 := godyf.NewString(string(idBytes))
	idString2 := godyf.NewString(dataHash)

	var idLine bytes.Buffer
	idLine.WriteString("/ID [")
	idLine.Write(idString1.Data())
	idLine.WriteString(" ")
	idLine.Write(idString2.Data())
	idLine.WriteString("]")

	return p.WriteLine(idLine.Bytes(), output)
}
