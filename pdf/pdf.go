package pdf

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"strings"

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

// Compressible returns whether the object can be included in an object stream
func (o *ObjectWrapper) Compressible() bool {
	return o.Object.Generation == 0
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

// writeCompressed writes PDF with compression
func (p *PDF) writeCompressed(output io.Writer, identifier interface{}) error {
	// Store compressed objects for later and write other ones in PDF
	var compressedObjects []godyf.PDFObject

	for _, obj := range p.Objects {
		objBase := obj.GetObject()
		if objBase.Free == 'f' {
			continue
		}

		if obj.Compressible() {
			compressedObjects = append(compressedObjects, obj)
		} else {
			objBase.Offset = p.CurrentPosition
			indirect := objBase.Indirect(obj.Data())
			if err := p.WriteLine(indirect, output); err != nil {
				return err
			}
		}
	}

	// Write compressed objects in object stream
	var stream []interface{}
	var offsetsAndNumbers []interface{}
	position := 0

	for _, obj := range compressedObjects {
		data := obj.Data()
		stream = append(stream, data)
		offsetsAndNumbers = append(offsetsAndNumbers, obj.GetObject().Number)
		offsetsAndNumbers = append(offsetsAndNumbers, position)
		position += len(data) + 1
	}

	// Create the first entry with object numbers and positions
	var firstEntry strings.Builder
	for i, item := range offsetsAndNumbers {
		if i > 0 {
			firstEntry.WriteString(" ")
		}
		firstEntry.WriteString(fmt.Sprintf("%v", item))
	}

	// Insert the first entry at the beginning
	streamData := make([]interface{}, 0, len(stream)+1)
	streamData = append(streamData, firstEntry.String())
	streamData = append(streamData, stream...)

	extra := map[string]interface{}{
		"Type":  "/ObjStm",
		"N":     len(compressedObjects),
		"First": len(firstEntry.String()) + 1,
	}

	objectStream := godyf.NewStream(streamData, extra, true)
	objectStream.GetObject().Offset = p.CurrentPosition
	p.AddObject(objectStream)

	indirect := objectStream.GetObject().Indirect(objectStream.Data())
	if err := p.WriteLine(indirect, output); err != nil {
		return err
	}

	// Write cross-reference stream
	var xref [][]int
	dictIndex := 0

	for _, obj := range p.Objects {
		objBase := obj.GetObject()
		if obj.Compressible() {
			xref = append(xref, []int{2, objectStream.GetObject().Number, dictIndex})
			dictIndex++
		} else {
			xref = append(xref, []int{
				boolToInt(objBase.Number != 0),
				objBase.Offset,
				objBase.Generation,
			})
		}
	}

	// Add entry for the current position
	xref = append(xref, []int{1, p.CurrentPosition, 0})

	// Calculate field sizes
	field2Size := int(math.Ceil(math.Log(float64(p.CurrentPosition+1)) / math.Log(256)))
	maxGeneration := 0
	for _, obj := range p.Objects {
		if obj.GetObject().Generation > maxGeneration {
			maxGeneration = obj.GetObject().Generation
		}
	}
	maxValue := maxGeneration
	if len(compressedObjects) > maxValue {
		maxValue = len(compressedObjects)
	}
	field3Size := int(math.Ceil(math.Log(float64(maxValue+1)) / math.Log(256)))

	xrefLengths := []int{1, field2Size, field3Size}

	// Create xref stream data
	var xrefStream bytes.Buffer
	for _, line := range xref {
		for i, value := range line {
			length := xrefLengths[i]
			data := make([]byte, length)
			// Convert to big-endian bytes
			for j := length - 1; j >= 0; j-- {
				data[j] = byte(value & 0xFF)
				value >>= 8
			}
			xrefStream.Write(data)
		}
	}

	extra = map[string]interface{}{
		"Type":  "/XRef",
		"Index": godyf.NewArray(0, len(p.Objects)+1),
		"W":     godyf.NewArray(xrefLengths[0], xrefLengths[1], xrefLengths[2]),
		"Size":  len(p.Objects) + 1,
		"Root":  string(p.Catalog.GetObject().Reference()),
		"Info":  string(p.Info.GetObject().Reference()),
	}

	if identifier != nil {
		if err := p.addIdentifierToExtra(extra, identifier); err != nil {
			return err
		}
	}

	dictStream := godyf.NewStream([]interface{}{xrefStream.Bytes()}, extra, true)
	p.XRefPosition = p.CurrentPosition
	dictStream.GetObject().Offset = p.CurrentPosition
	p.AddObject(dictStream)

	indirect = dictStream.GetObject().Indirect(dictStream.Data())
	if err := p.WriteLine(indirect, output); err != nil {
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

// boolToInt converts boolean to int (1 for true, 0 for false)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// addIdentifierToExtra adds identifier to the extra dictionary
func (p *PDF) addIdentifierToExtra(extra map[string]interface{}, identifier interface{}) error {
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

	extra["ID"] = godyf.NewArray(string(idString1.Data()), string(idString2.Data()))
	return nil
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
