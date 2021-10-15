package data

import (
	"encoding/binary"
	"math"
)

type DataProvider interface{} {
	RegisterWithBufferService(chan []byte)
	ServiceName() string
}

func Encode(value interface{}) []byte {
	var out []byte
	switch v := value.(type) {
	case string:
		out = append(out, byte(STRING_FIELD))
		out = append(out, []byte(v)...)
	case float64:
		out = append(out, byte(FLOAT64_FIELD))
		tmp := make([]byte, 8)
		binary.BigEndian.PutUint64(tmp, math.Float64bits(v))
		out = append(out, tmp...)
	}
	return out
}

func Decode(dataBytes []byte) (*DataFrame, error) {
	var timestamp string
	var data map[int64]DataField
	if dataBytes[0] != STRING_FIELD {
		return nil, fmt.Errorf("Expected string as first field. Received: %v", dataBytes[0])
	}
	timestamp = string(dataBytes[1:9])
	var j int64
	var floatValue float64
	var stringValue string
	var tag string
	for i := 9; i < len(dataBytes) {
		switch i {
		case STRING_FIELD:
		case FLOAT64_FIELD:
			floatValue = binary.BigEndian.Uint64(dataBytes[i+1:i+9])
			i += 9
		}
	}
}