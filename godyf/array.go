package godyf

import (
	"bytes"
)

// Array represents a PDF Array object
type Array struct {
	Object
	Elements []interface{} // Array elements
}

// NewArray creates a new Array object with the given elements
func NewArray(elements ...interface{}) *Array {
	return &Array{
		Object:   *NewObject(),
		Elements: elements,
	}
}

// NewArrayFromSlice creates a new Array object from a slice
func NewArrayFromSlice(slice []interface{}) *Array {
	return &Array{
		Object:   *NewObject(),
		Elements: slice,
	}
}

// Data returns the PDF representation of the array
func (a *Array) Data() []byte {
	var buf bytes.Buffer
	buf.WriteByte('[')

	// Convert each element to bytes and join with spaces
	for i, element := range a.Elements {
		if i > 0 {
			buf.WriteByte(' ')
		}
		buf.Write(ToBytes(element))
	}

	buf.WriteByte(']')
	return buf.Bytes()
}

// Add appends an element to the array
func (a *Array) Add(element interface{}) {
	a.Elements = append(a.Elements, element)
}

// Get returns the element at the given index
func (a *Array) Get(index int) interface{} {
	if index < 0 || index >= len(a.Elements) {
		return nil
	}
	return a.Elements[index]
}

// Len returns the number of elements in the array
func (a *Array) Len() int {
	return len(a.Elements)
}

// Set sets the element at the given index
func (a *Array) Set(index int, element interface{}) {
	if index >= 0 && index < len(a.Elements) {
		a.Elements[index] = element
	}
}

// GetObject returns the underlying Object struct
func (a *Array) GetObject() *Object {
	return &a.Object
}

// SetObject sets the underlying Object struct
func (a *Array) SetObject(obj *Object) {
	a.Object = *obj
}

// Compressible returns whether the array can be included in an object stream
func (a *Array) Compressible() bool {
	return a.Object.Generation == 0
}
