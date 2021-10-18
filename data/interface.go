package data

import (
	"encoding/binary"
	"fmt"
	"math"
)

type DataProvider interface {
	RegisterWithBufferService(*DataBuffer) error
	ServiceName() string
}

// Encode encodes different data types into bytes using BigEndian byte order
func Encode(value interface{}) []byte {
	var out []byte
	switch v := value.(type) {
	case string:
		out = append(out, byte(STRING_FIELD))
		tmp := make([]byte, 8)
		binary.BigEndian.PutUint64(tmp, uint64(int64(len(v))))
		out = append(out, tmp...)
		out = append(out, []byte(v)...)
	case float64:
		out = append(out, byte(FLOAT64_FIELD))
		tmp := make([]byte, 8)
		binary.BigEndian.PutUint64(tmp, uint64(int64(8)))
		out = append(out, tmp...)
		binary.BigEndian.PutUint64(tmp, math.Float64bits(v))
		out = append(out, tmp...)
	}
	return out
}

// Decode decodes data from a DataProducer and converts it into a DataFrame struct.
// The decoding process follows this order:
// - first byte is data type (should be a string)
// - The next 8 bytes will be length of the incoming data (int64)
// - Then the following X bytes are the value
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
	var lengthValue uint64
	for i := 9; i < len(dataBytes); {
		switch dataBytes[i] {
		case STRING_FIELD:
			i++
			lengthValue = binary.BigEndian.Uint64(dataBytes[i:i+8])
			i += 8
			stringValue = string(dataBytes[i:i+int(lengthValue)])
			data[j] = DataField{
				Name: stringValue,
			}
			i += int(lengthValue)
		case FLOAT64_FIELD:
			i++
			lengthValue = binary.BigEndian.Uint64(dataBytes[i:i+8])
			i += 8
			tmp := binary.BigEndian.Uint64(dataBytes[i:i+int(lengthValue)])
			floatValue = math.Float64frombits(tmp)
			fieldCopy := data[j]
			fieldCopy.Value = floatValue
			data[j] = fieldCopy
			i += int(lengthValue)
			j++
		}
	}
	return &DataFrame{
		IsScanning: true,
		Timestamp: timestamp,
		Data: data,
	}, nil
}