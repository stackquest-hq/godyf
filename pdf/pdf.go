package pdf

import (
	"crypto/md5"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/stackquest-hq/godyf/godyf"
)

// PDF represents a PDF document
type PDF struct {
	Objects         []godyf.PDFObject // List containing the PDF's objects
	Pages           *godyf.Dictionary // Dictionary containing the PDF's pages
	Info            *godyf.Dictionary // Dictionary containing the PDF's metadata
	Catalog         *godyf.Dictionary // Dictionary containing references to other objects
	CurrentPosition int               // Current position in the PDF
	XrefPosition    *int              // Position of the cross reference table
}

// NewPDF creates a new PDF document
func NewPDF() *PDF {
	pdf := &PDF{
		Objects:         make([]godyf.PDFObject, 0),
		CurrentPosition: 0,
		XrefPosition:    nil,
	}

	// Create zero object
	zeroObject := godyf.NewObject()
	zeroObject.Generation = 65535
	zeroObject.Free = 'f'
	pdf.AddObject(zeroObject)

	// Create pages dictionary
	pdf.Pages = godyf.NewDictionary(map[string]interface{}{
		"Type":  "/Pages",
		"Kids":  godyf.NewArray(),
		"Count": 0,
	})
	pdf.AddObject(pdf.Pages)

	// Create info dictionary
	pdf.Info = godyf.NewDictionary(map[string]interface{}{})
	pdf.AddObject(pdf.Info)

	// Create catalog dictionary
	pdf.Catalog = godyf.NewDictionary(map[string]interface{}{
		"Type":  "/Catalog",
		"Pages": string(pdf.Pages.Refrence()), // Use Object's reference method
	})
	pdf.AddObject(pdf.Catalog)

	return pdf
}

// AddPage adds a page to the PDF
func (p *PDF) AddPage(page *godyf.Dictionary) {
	// Increment page count
	if count, ok := p.Pages.Values["Count"].(int); ok {
		p.Pages.Values["Count"] = count + 1
	}

	// Add page object
	p.AddObject(page)

	// Add page reference to Kids array
	if kids, ok := p.Pages.Values["Kids"].(*godyf.Array); ok {
		kids.Add(page.GetNumber())
		kids.Add(0)
		kids.Add("R")
	}
}

// AddObject adds an object to the PDF
func (p *PDF) AddObject(obj godyf.PDFObject) {
	obj.SetNumber(len(p.Objects))
	p.Objects = append(p.Objects, obj)
}

// PageReferences returns a tuple of page references
func (p *PDF) PageReferences() []string {
	var refs []string
	if kids, ok := p.Pages.Values["Kids"].(*godyf.Array); ok {
		for i := 0; i < len(kids.Elements); i += 3 {
			if objNum, ok := kids.Elements[i].(int); ok {
				refs = append(refs, fmt.Sprintf("%d 0 R", objNum))
			}
		}
	}
	return refs
}

// WriteLine writes a line to the output and updates current position
func (p *PDF) WriteLine(content []byte, output io.Writer) error {
	p.CurrentPosition += len(content) + 1
	_, err := output.Write(content)
	if err != nil {
		return err
	}
	_, err = output.Write([]byte{'\n'})
	return err
}

// Write writes the PDF to output
func (p *PDF) Write(output io.Writer, version string, identifier interface{}, compress bool) error {
	// Convert version
	if version == "" {
		version = "1.7"
	}
	versionBytes := []byte(version)

	// Handle identifier
	var identifierBytes []byte
	var generateIdentifier bool
	switch id := identifier.(type) {
	case bool:
		generateIdentifier = id
	case string:
		identifierBytes = []byte(id)
	case []byte:
		identifierBytes = id
	default:
		generateIdentifier = false
	}

	// Write header
	if err := p.WriteLine(append([]byte("%PDF-"), versionBytes...), output); err != nil {
		return err
	}
	if err := p.WriteLine([]byte{0x25, 0xf0, 0x9f, 0x96, 0xa4}, output); err != nil {
		return err
	}

	if version >= "1.5" && compress {
		return p.writeCompressed(output, identifierBytes, generateIdentifier)
	} else {
		return p.writeUncompressed(output, identifierBytes, generateIdentifier)
	}
}

// writeUncompressed writes PDF in uncompressed format
func (p *PDF) writeUncompressed(output io.Writer, identifierBytes []byte, generateIdentifier bool) error {
	// Write all non-free PDF objects
	for _, obj := range p.Objects {
		if obj.GetFree() == 'f' {
			continue
		}
		obj.SetOffset(p.CurrentPosition)
		if err := p.WriteLine(obj.Indirect(), output); err != nil {
			return err
		}
	}

	// Write cross-reference table
	xrefPos := p.CurrentPosition
	p.XrefPosition = &xrefPos

	if err := p.WriteLine([]byte("xref"), output); err != nil {
		return err
	}
	if err := p.WriteLine([]byte(fmt.Sprintf("0 %d", len(p.Objects))), output); err != nil {
		return err
	}

	for _, obj := range p.Objects {
		line := fmt.Sprintf("%010d %05d %c ", obj.GetOffset(), obj.GetGeneration(), obj.GetFree())
		if err := p.WriteLine([]byte(line), output); err != nil {
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
	if err := p.WriteLine([]byte(fmt.Sprintf("/Size %d", len(p.Objects))), output); err != nil {
		return err
	}
	if err := p.WriteLine(append([]byte("/Root "), p.Catalog.Refrence()...), output); err != nil {
		return err
	}
	if err := p.WriteLine(append([]byte("/Info "), p.Info.Refrence()...), output); err != nil {
		return err
	}

	// Handle identifier
	if generateIdentifier || identifierBytes != nil {
		var dataHash []byte
		if generateIdentifier {
			// Generate hash from all object data
			hasher := md5.New()
			for _, obj := range p.Objects {
				if obj.GetFree() != 'f' {
					hasher.Write(obj.Data())
				}
			}
			dataHash = []byte(fmt.Sprintf("%x", hasher.Sum(nil)))
			if identifierBytes == nil {
				identifierBytes = dataHash
			}
		} else {
			// Generate hash for second identifier
			hasher := md5.New()
			for _, obj := range p.Objects {
				if obj.GetFree() != 'f' {
					hasher.Write(obj.Data())
				}
			}
			dataHash = []byte(fmt.Sprintf("%x", hasher.Sum(nil)))
		}

		id1 := godyf.NewString(string(identifierBytes))
		id2 := godyf.NewString(string(dataHash))

		idLine := append([]byte("/ID ["), id1.Data()...)
		idLine = append(idLine, ' ')
		idLine = append(idLine, id2.Data()...)
		idLine = append(idLine, ']')

		if err := p.WriteLine(idLine, output); err != nil {
			return err
		}
	}

	if err := p.WriteLine([]byte(">>"), output); err != nil {
		return err
	}

	// Write startxref and EOF
	if err := p.WriteLine([]byte("startxref"), output); err != nil {
		return err
	}
	if err := p.WriteLine([]byte(fmt.Sprintf("%d", *p.XrefPosition)), output); err != nil {
		return err
	}
	if err := p.WriteLine([]byte("%%EOF"), output); err != nil {
		return err
	}

	return nil
}

// writeCompressed writes PDF in compressed format with object streams
func (p *PDF) writeCompressed(output io.Writer, identifierBytes []byte, generateIdentifier bool) error {
	// Store compressed objects for later and write other ones in PDF
	var compressedObjects []godyf.PDFObject

	for _, obj := range p.Objects {
		if obj.GetFree() == 'f' {
			continue
		}
		if obj.Compressible() {
			compressedObjects = append(compressedObjects, obj)
		} else {
			obj.SetOffset(p.CurrentPosition)
			if err := p.WriteLine(obj.Indirect(), output); err != nil {
				return err
			}
		}
	}

	// Write compressed objects in object stream
	stream := []interface{}{[]string{}}
	position := 0

	for _, obj := range compressedObjects {
		data := obj.Data()
		stream = append(stream, data)

		// Add to index
		if index, ok := stream[0].([]string); ok {
			index = append(index, strconv.Itoa(obj.GetNumber()))
			index = append(index, strconv.Itoa(position))
			stream[0] = index
		}
		position += len(data) + 1
	}

	// Convert index to string
	if index, ok := stream[0].([]string); ok {
		stream[0] = strings.Join(index, " ")
	}

	extra := map[string]interface{}{
		"Type":  "/ObjStm",
		"N":     len(compressedObjects),
		"First": len(stream[0].(string)) + 1,
	}

	objectStream := godyf.NewStream(stream, extra, true)
	objectStream.Offset = p.CurrentPosition
	p.AddObject(&objectStream.Object)
	if err := p.WriteLine(objectStream.Indirect(), output); err != nil {
		return err
	}

	// Write cross-reference stream
	var xref [][]int
	dictIndex := 0

	for _, obj := range p.Objects {
		if obj.Compressible() {
			xref = append(xref, []int{2, objectStream.Number, dictIndex})
			dictIndex++
		} else {
			status := 0
			if obj.GetNumber() > 0 {
				status = 1
			}
			xref = append(xref, []int{status, obj.GetOffset(), obj.GetGeneration()})
		}
	}
	xref = append(xref, []int{1, p.CurrentPosition, 0})

	// Calculate field sizes
	field2Size := int(math.Ceil(math.Log(float64(p.CurrentPosition+1)) / math.Log(256)))
	maxGeneration := 0
	for _, obj := range p.Objects {
		if obj.GetGeneration() > maxGeneration {
			maxGeneration = obj.GetGeneration()
		}
	}
	maxVal := maxGeneration
	if len(compressedObjects) > maxVal {
		maxVal = len(compressedObjects)
	}
	field3Size := int(math.Ceil(math.Log(float64(maxVal+1)) / math.Log(256)))

	xrefLengths := []int{1, field2Size, field3Size}

	// Create xref stream data
	var xrefStream []byte
	for _, line := range xref {
		for i, value := range line {
			length := xrefLengths[i]
			for j := length - 1; j >= 0; j-- {
				xrefStream = append(xrefStream, byte((value>>(8*j))&0xFF))
			}
		}
	}

	extra = map[string]interface{}{
		"Type":  "/XRef",
		"Index": godyf.NewArray(0, len(p.Objects)+1),
		"W":     godyf.NewArrayFromSlice([]interface{}{xrefLengths[0], xrefLengths[1], xrefLengths[2]}),
		"Size":  len(p.Objects) + 1,
		"Root":  string(p.Catalog.Refrence()),
		"Info":  string(p.Info.Refrence()),
	}

	// Handle identifier for compressed format
	if generateIdentifier || identifierBytes != nil {
		var dataHash []byte
		hasher := md5.New()
		for _, obj := range p.Objects {
			if obj.GetFree() != 'f' {
				hasher.Write(obj.Data())
			}
		}
		dataHash = []byte(fmt.Sprintf("%x", hasher.Sum(nil)))

		var finalIdentifierBytes []byte
		if generateIdentifier && identifierBytes == nil {
			finalIdentifierBytes = dataHash
		} else {
			finalIdentifierBytes = identifierBytes
		}

		id1 := godyf.NewString(string(finalIdentifierBytes))
		id2 := godyf.NewString(string(dataHash))
		extra["ID"] = godyf.NewArray(string(id1.Data()), string(id2.Data()))
	}

	dictStream := godyf.NewStream([]interface{}{xrefStream}, extra, true)
	xrefPos := p.CurrentPosition
	p.XrefPosition = &xrefPos
	dictStream.Offset = p.CurrentPosition
	p.AddObject(&dictStream.Object)
	if err := p.WriteLine(dictStream.Indirect(), output); err != nil {
		return err
	}

	// Write startxref and EOF
	if err := p.WriteLine([]byte("startxref"), output); err != nil {
		return err
	}
	if err := p.WriteLine([]byte(fmt.Sprintf("%d", *p.XrefPosition)), output); err != nil {
		return err
	}
	if err := p.WriteLine([]byte("%%EOF"), output); err != nil {
		return err
	}

	return nil
}
