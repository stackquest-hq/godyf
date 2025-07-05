package godyf

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
)

// Stream represents a PDF Stream object
type Stream struct {
	Object
	Stream   []interface{}          // Array of data composing stream
	Extra    map[string]interface{} // Metadata containing at least the length of the Stream
	Compress bool                   // Compress the stream data if set to true
}

// NewStream creates a new Stream object
func NewStream(stream []interface{}, extra map[string]interface{}, compress bool) *Stream {
	if stream == nil {
		stream = make([]interface{}, 0)
	}
	if extra == nil {
		extra = make(map[string]interface{})
	}
	return &Stream{
		Object:   *NewObject(),
		Stream:   stream,
		Extra:    extra,
		Compress: compress,
	}
}

// Compressible overrides Object.Compressible to return false for streams
// Stream objects cannot be included in object streams according to PDF spec
func (s *Stream) Compressible() bool {
	return false
}

// Data returns the PDF representation of the stream
func (s *Stream) Data() []byte {
	// Join all stream elements with newlines
	var streamData bytes.Buffer
	for i, item := range s.Stream {
		if i > 0 {
			streamData.WriteByte('\n')
		}
		streamData.Write(ToBytes(item))
	}

	stream := streamData.Bytes()
	extra := make(map[string]interface{})

	// Copy extra metadata
	for k, v := range s.Extra {
		extra[k] = v
	}

	// Compress if requested
	if s.Compress {
		extra["Filter"] = "/FlateDecode"
		var buf bytes.Buffer
		writer := zlib.NewWriter(&buf)
		writer.Write(stream)
		writer.Close()
		stream = buf.Bytes()
	}

	// Set length
	extra["Length"] = len(stream)

	// Create dictionary for extra data
	dict := NewDictionary(extra)

	// Build final stream data
	var result bytes.Buffer
	result.Write(dict.Data())
	result.WriteString("\nstream\n")
	result.Write(stream)
	result.WriteString("\nendstream")

	return result.Bytes()
}

// BeginMarkedContent begins marked-content sequence
func (s *Stream) BeginMarkedContent(tag string, propertyList interface{}) {
	s.Stream = append(s.Stream, "/"+tag)
	if propertyList == nil {
		s.Stream = append(s.Stream, "BMC")
	} else {
		s.Stream = append(s.Stream, propertyList)
		s.Stream = append(s.Stream, "BDC")
	}
}

// BeginText begins a text object
func (s *Stream) BeginText() {
	s.Stream = append(s.Stream, "BT")
}

// Clip modifies current clipping path by intersecting it with current path
func (s *Stream) Clip(evenOdd bool) {
	if evenOdd {
		s.Stream = append(s.Stream, "W*")
	} else {
		s.Stream = append(s.Stream, "W")
	}
}

// Close closes current subpath
func (s *Stream) Close() {
	s.Stream = append(s.Stream, "h")
}

// CurveTo adds cubic Bézier curve to current path
func (s *Stream) CurveTo(x1, y1, x2, y2, x3, y3 float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s %s %s %s %s c",
		ToBytes(x1), ToBytes(y1), ToBytes(x2), ToBytes(y2), ToBytes(x3), ToBytes(y3)))
}

// CurveStartTo adds cubic Bézier curve to current path using current point as start
func (s *Stream) CurveStartTo(x2, y2, x3, y3 float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s %s %s v",
		ToBytes(x2), ToBytes(y2), ToBytes(x3), ToBytes(y3)))
}

// CurveEndTo adds cubic Bézier curve to current path
func (s *Stream) CurveEndTo(x1, y1, x3, y3 float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s %s %s y",
		ToBytes(x1), ToBytes(y1), ToBytes(x3), ToBytes(y3)))
}

// DrawXObject draws object given by reference
func (s *Stream) DrawXObject(reference string) {
	s.Stream = append(s.Stream, "/"+reference+" Do")
}

// End ends path without filling or stroking
func (s *Stream) End() {
	s.Stream = append(s.Stream, "n")
}

// EndMarkedContent ends marked-content sequence
func (s *Stream) EndMarkedContent() {
	s.Stream = append(s.Stream, "EMC")
}

// EndText ends text object
func (s *Stream) EndText() {
	s.Stream = append(s.Stream, "ET")
}

// Fill fills path using nonzero winding rule
func (s *Stream) Fill(evenOdd bool) {
	if evenOdd {
		s.Stream = append(s.Stream, "f*")
	} else {
		s.Stream = append(s.Stream, "f")
	}
}

// FillAndStroke fills and strokes path using nonzero winding rule
func (s *Stream) FillAndStroke(evenOdd bool) {
	if evenOdd {
		s.Stream = append(s.Stream, "B*")
	} else {
		s.Stream = append(s.Stream, "B")
	}
}

// FillStrokeAndClose fills, strokes and closes path using nonzero winding rule
func (s *Stream) FillStrokeAndClose(evenOdd bool) {
	if evenOdd {
		s.Stream = append(s.Stream, "b*")
	} else {
		s.Stream = append(s.Stream, "b")
	}
}

// InlineImage adds an inline image
func (s *Stream) InlineImage(width, height int, colorSpace string, bpc int, rawData []byte) {
	var data []byte
	if s.Compress {
		var buf bytes.Buffer
		writer := zlib.NewWriter(&buf)
		writer.Write(rawData)
		writer.Close()
		data = buf.Bytes()
	} else {
		data = rawData
	}

	// ASCII85 encode the data
	a85Data := base64.StdEncoding.EncodeToString(data) + "~>"

	var filter string
	if s.Compress {
		filter = "[/A85 /Fl]"
	} else {
		filter = "/A85"
	}

	imageCmd := fmt.Sprintf("BI /W %d /H %d /BPC %d /CS /Device%s /F %s /L %d ID %s EI",
		width, height, bpc, colorSpace, filter, len(a85Data), a85Data)
	s.Stream = append(s.Stream, imageCmd)
}

// LineTo adds line from current point to point (x, y)
func (s *Stream) LineTo(x, y float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s l", ToBytes(x), ToBytes(y)))
}

// MoveTo begins new subpath by moving current point to (x, y)
func (s *Stream) MoveTo(x, y float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s m", ToBytes(x), ToBytes(y)))
}

// MoveTextTo moves text to next line at (x, y) distance from previous line
func (s *Stream) MoveTextTo(x, y float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s Td", ToBytes(x), ToBytes(y)))
}

// PaintShading paints shape and color shading using shading dictionary name
func (s *Stream) PaintShading(name string) {
	s.Stream = append(s.Stream, "/"+name+" sh")
}

// PopState restores graphic state
func (s *Stream) PopState() {
	s.Stream = append(s.Stream, "Q")
}

// PushState saves graphic state
func (s *Stream) PushState() {
	s.Stream = append(s.Stream, "q")
}

// Rectangle adds rectangle to current path as complete subpath
func (s *Stream) Rectangle(x, y, width, height float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s %s %s re",
		ToBytes(x), ToBytes(y), ToBytes(width), ToBytes(height)))
}

// SetColorRGB sets RGB color for nonstroking operations
func (s *Stream) SetColorRGB(r, g, b float64, stroke bool) {
	var op string
	if stroke {
		op = "RG"
	} else {
		op = "rg"
	}
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s %s %s", ToBytes(r), ToBytes(g), ToBytes(b), op))
}

// SetColorSpace sets the nonstroking color space
func (s *Stream) SetColorSpace(space string, stroke bool) {
	var op string
	if stroke {
		op = "CS"
	} else {
		op = "cs"
	}
	s.Stream = append(s.Stream, "/"+space+" "+op)
}

// SetColorSpecial sets special color for nonstroking operations
func (s *Stream) SetColorSpecial(name string, stroke bool, operands ...interface{}) {
	var parts []string
	for _, operand := range operands {
		parts = append(parts, string(ToBytes(operand)))
	}
	if name != "" {
		parts = append(parts, "/"+name)
	}

	var op string
	if stroke {
		op = "SCN"
	} else {
		op = "scn"
	}

	var buf bytes.Buffer
	for i, part := range parts {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.WriteString(part)
	}
	buf.WriteByte(' ')
	buf.WriteString(op)

	s.Stream = append(s.Stream, buf.String())
}

// SetDash sets dash line pattern
func (s *Stream) SetDash(dashArray []float64, dashPhase float64) {
	array := NewArrayFromSlice(make([]interface{}, len(dashArray)))
	for i, val := range dashArray {
		array.Elements[i] = val
	}
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s d", array.Data(), ToBytes(dashPhase)))
}

// SetFontSize sets font name and size
func (s *Stream) SetFontSize(font string, size float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("/%s %s Tf", font, ToBytes(size)))
}

// SetTextRendering sets text rendering mode
func (s *Stream) SetTextRendering(mode int) {
	s.Stream = append(s.Stream, fmt.Sprintf("%d Tr", mode))
}

// SetTextRise sets text rise
func (s *Stream) SetTextRise(height float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s Ts", ToBytes(height)))
}

// SetLineCap sets line cap style
func (s *Stream) SetLineCap(lineCap int) {
	s.Stream = append(s.Stream, fmt.Sprintf("%d J", lineCap))
}

// SetLineJoin sets line join style
func (s *Stream) SetLineJoin(lineJoin int) {
	s.Stream = append(s.Stream, fmt.Sprintf("%d j", lineJoin))
}

// SetLineWidth sets line width
func (s *Stream) SetLineWidth(width float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s w", ToBytes(width)))
}

// SetMatrix sets current transformation matrix
func (s *Stream) SetMatrix(a, b, c, d, e, f float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s %s %s %s %s cm",
		ToBytes(a), ToBytes(b), ToBytes(c), ToBytes(d), ToBytes(e), ToBytes(f)))
}

// SetMiterLimit sets miter limit
func (s *Stream) SetMiterLimit(miterLimit float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s M", ToBytes(miterLimit)))
}

// SetState sets specified parameters in graphic state
func (s *Stream) SetState(stateName string) {
	s.Stream = append(s.Stream, "/"+stateName+" gs")
}

// SetTextMatrix sets current text and text line transformation matrix
func (s *Stream) SetTextMatrix(a, b, c, d, e, f float64) {
	s.Stream = append(s.Stream, fmt.Sprintf("%s %s %s %s %s %s Tm",
		ToBytes(a), ToBytes(b), ToBytes(c), ToBytes(d), ToBytes(e), ToBytes(f)))
}

// ShowText shows text strings with individual glyph positioning
func (s *Stream) ShowText(text string) {
	s.Stream = append(s.Stream, fmt.Sprintf("[%s] TJ", ToBytes(text)))
}

// ShowTextString shows single text string
func (s *Stream) ShowTextString(text string) {
	str := NewString(text)
	s.Stream = append(s.Stream, fmt.Sprintf("%s Tj", str.Data()))
}

// Stroke strokes path
func (s *Stream) Stroke() {
	s.Stream = append(s.Stream, "S")
}

// StrokeAndClose strokes and closes path
func (s *Stream) StrokeAndClose() {
	s.Stream = append(s.Stream, "s")
}
