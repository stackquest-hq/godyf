package godyf

import (
	"fmt"
)

// Object represents a PDF object with number, offset, generation, and free status
type Object struct {
	Number     int
	Offset     int
	Generation int
	Free       rune
}

// PDFObject interface defines methods that all PDF objects must implement
type PDFObject interface {
	Data() []byte
	GetObject() *Object
	SetObject(*Object)
}

// NewObject creates a new PDF object with default values
func NewObject() *Object {
	return &Object{
		Number:     0,
		Offset:     0,
		Generation: 0,
		Free:       'n',
	}
}

// Indirect returns the indirect representation of an object
func (o *Object) Indirect(data []byte) []byte {
	header := fmt.Sprintf("%d %d obj\n", o.Number, o.Generation)
	result := []byte(header)
	result = append(result, data...)
	result = append(result, []byte("\nendobj")...)
	return result
}

// Reference returns the object identifier
func (o *Object) Reference() []byte {
	return []byte(fmt.Sprintf("%d %d R", o.Number, o.Generation))
}

// Compressible returns whether the object can be included in an object stream
func (o *Object) Compressible(obj interface{}) bool {
	if o.Generation != 0 {
		return false
	}
	if _, isStream := obj.(*Stream); isStream {
		return false
	}
	return true
}
