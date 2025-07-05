package godyf

import (
	"bytes"
	"fmt"
	"reflect"
)

// PDFObject interface defines methods that all PDF objects must implement
type PDFObject interface {
	Data() []byte
	GetNumber() int
	SetNumber(int)
	GetGeneration() int
	SetGeneration(int)
	GetOffset() int
	SetOffset(int)
	GetFree() rune
	SetFree(rune)
	Indirect() []byte
	Refrence() []byte
	Compressible() bool
}

type Object struct {
	Number     int  //Number of the object
	Offset     int  //Position in the pdf of the object
	Generation int  // version number of the object, non-negative
	Free       rune // Indicates if an object is used (`'n'`), or has been deleted and therefor is free (`'f'`)
}

// return a new Object with default values
func NewObject() *Object {
	return &Object{
		Number:     0,
		Offset:     0,
		Generation: 0,
		Free:       'n',
	}
}

// Data should be implemented by concrete types
// For the base Object (typically the zero object), return empty data
func (o *Object) Data() []byte {
	// Zero object (object 0) typically has no data
	if o.Number == 0 {
		return []byte{}
	}
	panic("Data method not implemented")
}

// Indirect representation of the object
// e.g. "1 0 obj ... endobj"
func (o *Object) Indirect() []byte {
	var buf bytes.Buffer
	buf.WriteString((fmt.Sprintf("%d %d obj\n", o.Number, o.Generation)))
	buf.Write(o.Data())
	buf.WriteString("\nendobj\n")
	return buf.Bytes()
}

// Object identifier
// e.g. "1 0 R"
func (o *Object) Refrence() []byte {
	return []byte(fmt.Sprintf("%d %d R", o.Number, o.Generation))
}

// Whether the object can be included in an Object stream
func (o *Object) Compressible() bool {
	// return o.Free == 'n' && o.Generation == 0
	if o.Generation != 0 || o.Free != 'n' {
		return false
	}
	objectType := reflect.TypeOf(o).Elem().Name()
	return objectType != "Stream"
}

// Getter and setter methods to satisfy the PDFObject interface
func (o *Object) GetNumber() int {
	return o.Number
}

func (o *Object) SetNumber(n int) {
	o.Number = n
}

func (o *Object) GetGeneration() int {
	return o.Generation
}

func (o *Object) SetGeneration(g int) {
	o.Generation = g
}

func (o *Object) GetOffset() int {
	return o.Offset
}

func (o *Object) SetOffset(offset int) {
	o.Offset = offset
}

func (o *Object) GetFree() rune {
	return o.Free
}

func (o *Object) SetFree(f rune) {
	o.Free = f
}
