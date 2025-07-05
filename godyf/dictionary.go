package godyf

import "bytes"

type Dictionary struct {
	Object // Embed the Object struct to inherit its properties
	Values map[string]interface{}
}

// NewDictionary creates a new Dictionary object with optional initial values.
func NewDictionary(values map[string]interface{}) *Dictionary {
	if values == nil {
		values = make(map[string]interface{})
	}
	return &Dictionary{
		Object: Object{
			Free: 'n',
		},
		Values: values,
	}
}

// Data returns the PDF byte representation of the dictionary.
func (d *Dictionary) Data() []byte {
	var buf bytes.Buffer
	buf.WriteString("<<")
	for key, value := range d.Values {
		buf.WriteString(" /")
		buf.Write(ToBytes(key))
		buf.WriteByte(' ')
		buf.Write(ToBytes(value))
	}
	buf.WriteString(" >>")
	return buf.Bytes()
}
