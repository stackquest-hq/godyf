package godyf

import (
	"bytes"
	"encoding/hex"
	"regexp"
	"unicode/utf16"
)

// String represents a PDF String object
type String struct {
	Object
	String string // Unicode string content
}

// NewString creates a new String object with the given string content
func NewString(s string) *String {
	return &String{
		Object: *NewObject(),
		String: s,
	}
}

// Data returns the PDF representation of the string
func (s *String) Data() []byte {
	// Try literal string encoding first (like Python's try-except approach)
	if s.canBeLiteralString() {
		return s.literalStringData()
	}
	// If it contains non-ASCII characters or can't be literal, use hex encoding
	return s.hexStringData()
}

// canBeLiteralString checks if the string can be represented as a literal string
func (s *String) canBeLiteralString() bool {
	// Check if string contains only ASCII characters and no problematic characters
	for _, r := range s.String {
		if r > 127 || r < 32 {
			return false
		}
	}
	return true
}

// literalStringData returns the literal string representation
func (s *String) literalStringData() []byte {
	var buf bytes.Buffer
	buf.WriteByte('(')

	// Convert string to bytes first, like the Python version
	stringBytes := ToBytes(s.String)

	// Escape special characters: backslash, parentheses (work with bytes like Python)
	re := regexp.MustCompile(`([\\()])`)
	escaped := re.ReplaceAll(stringBytes, []byte(`\$1`))

	buf.Write(escaped)
	buf.WriteByte(')')
	return buf.Bytes()
}

// hexStringData returns the hex-encoded string representation
func (s *String) hexStringData() []byte {
	var buf bytes.Buffer
	buf.WriteByte('<')

	// Add BOM for UTF-16 BE and encode string
	utf16Bytes := utf16.Encode([]rune(s.String))

	// Convert to bytes in big-endian format with BOM
	var encoded bytes.Buffer
	encoded.Write([]byte{0xFE, 0xFF}) // BOM for UTF-16 BE

	for _, r := range utf16Bytes {
		encoded.WriteByte(byte(r >> 8))   // High byte
		encoded.WriteByte(byte(r & 0xFF)) // Low byte
	}

	// Convert to hex
	hexStr := hex.EncodeToString(encoded.Bytes())
	buf.WriteString(hexStr)
	buf.WriteByte('>')
	return buf.Bytes()
}

// GetObject returns the underlying Object struct
func (s *String) GetObject() *Object {
	return &s.Object
}

// SetObject sets the underlying Object struct
func (s *String) SetObject(obj *Object) {
	s.Object = *obj
}

// Compressible returns whether the string can be included in an object stream
func (s *String) Compressible() bool {
	return s.Object.Generation == 0
}
