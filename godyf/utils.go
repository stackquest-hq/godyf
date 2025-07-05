package godyf

import (
	"fmt"
	"strconv"
	"strings"
)

type ByteData interface {
	Data() []byte
}

func ToBytes(item interface{}) []byte {
	switch v := item.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	case float64:
		if v == float64(int64(v)) {
			return []byte(strconv.FormatInt(int64(v), 10))
		} else {
			s := strconv.FormatFloat(v, 'f', -1, 64)
			s = strings.TrimRight(s, "0")
			if strings.HasSuffix(s, ".") {
				s += "0"
			}
			return []byte(s)
		}
	case ByteData:
		return v.Data()
	default:
		return []byte(fmt.Sprintf("%v", v))
	}
}
