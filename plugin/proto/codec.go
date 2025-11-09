package proto

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
)

const (
	wireVarint     = 0
	wireFixed64    = 1
	wireLength     = 2
	wireStartGroup = 3
	wireEndGroup   = 4
	wireFixed32    = 5
)

func appendVarint(buf []byte, v uint64) []byte {
	for v >= 0x80 {
		buf = append(buf, byte(v)|0x80)
		v >>= 7
	}
	return append(buf, byte(v))
}

func appendTag(buf []byte, field int, wire int) []byte {
	return appendVarint(buf, uint64(field<<3|wire))
}

func appendString(buf []byte, field int, value string) []byte {
	if value == "" {
		return buf
	}
	buf = appendTag(buf, field, wireLength)
	buf = appendVarint(buf, uint64(len(value)))
	return append(buf, value...)
}

func appendBytes(buf []byte, field int, value []byte) []byte {
	if len(value) == 0 {
		return buf
	}
	buf = appendTag(buf, field, wireLength)
	buf = appendVarint(buf, uint64(len(value)))
	return append(buf, value...)
}

func decodeVarint(data []byte) (uint64, int, error) {
	var value uint64
	for i := 0; i < len(data); i++ {
		b := data[i]
		if i == 9 && b > 1 {
			return 0, 0, errors.New("varint overflow")
		}
		value |= uint64(b&0x7F) << (7 * i)
		if b < 0x80 {
			return value, i + 1, nil
		}
	}
	return 0, 0, errors.New("unexpected end of varint")
}

func readTag(data []byte) (field int, wire int, n int, err error) {
	v, consumed, err := decodeVarint(data)
	if err != nil {
		return 0, 0, 0, err
	}
	field = int(v >> 3)
	wire = int(v & 0x7)
	if field == 0 {
		return 0, 0, 0, fmt.Errorf("invalid field number")
	}
	return field, wire, consumed, nil
}

func readString(data []byte) (string, int, error) {
	l, n, err := decodeVarint(data)
	if err != nil {
		return "", 0, err
	}
	if int(l) < 0 || int(l) > len(data[n:]) {
		return "", 0, errors.New("invalid length-delimited field")
	}
	start := n
	end := n + int(l)
	return string(data[start:end]), end, nil
}

func readBytes(data []byte) ([]byte, int, error) {
	l, n, err := decodeVarint(data)
	if err != nil {
		return nil, 0, err
	}
	if int(l) < 0 || int(l) > len(data[n:]) {
		return nil, 0, errors.New("invalid length-delimited field")
	}
	start := n
	end := n + int(l)
	return data[start:end], end, nil
}

func skipField(data []byte, wire int) (int, error) {
	switch wire {
	case wireVarint:
		_, n, err := decodeVarint(data)
		return n, err
	case wireFixed64:
		if len(data) < 8 {
			return 0, errors.New("unexpected EOF skipping fixed64")
		}
		return 8, nil
	case wireLength:
		l, n, err := decodeVarint(data)
		if err != nil {
			return 0, err
		}
		if int(l) < 0 || int(l) > len(data[n:]) {
			return 0, errors.New("invalid length-delimited field")
		}
		return n + int(l), nil
	case wireFixed32:
		if len(data) < 4 {
			return 0, errors.New("unexpected EOF skipping fixed32")
		}
		return 4, nil
	default:
		return 0, fmt.Errorf("unsupported wire type %d", wire)
	}
}

func appendMessage(buf []byte, field int, msg marshalable) ([]byte, error) {
	if msg == nil {
		return buf, nil
	}
	b, err := msg.Marshal()
	if err != nil {
		return nil, err
	}
	buf = appendTag(buf, field, wireLength)
	buf = appendVarint(buf, uint64(len(b)))
	buf = append(buf, b...)
	return buf, nil
}

type marshalable interface {
	Marshal() ([]byte, error)
}

func readMessage(data []byte, target unmarshallable) (int, error) {
	b, n, err := readBytes(data)
	if err != nil {
		return 0, err
	}
	if err := target.Unmarshal(b); err != nil {
		return 0, err
	}
	return n, nil
}

type unmarshallable interface {
	Unmarshal([]byte) error
}

func appendBool(buf []byte, field int, value bool) []byte {
	if !value {
		return buf
	}
	buf = appendTag(buf, field, wireVarint)
	if value {
		return append(buf, 1)
	}
	return append(buf, 0)
}

func readBool(data []byte) (bool, int, error) {
	v, n, err := decodeVarint(data)
	if err != nil {
		return false, 0, err
	}
	return v != 0, n, nil
}

func appendFloat32(buf []byte, field int, value float32) []byte {
	if value == 0 {
		return buf
	}
	buf = appendTag(buf, field, wireFixed32)
	var scratch [4]byte
	binary.LittleEndian.PutUint32(scratch[:], math.Float32bits(value))
	return append(buf, scratch[:]...)
}

func appendFloat64(buf []byte, field int, value float64) []byte {
	if value == 0 {
		return buf
	}
	buf = appendTag(buf, field, wireFixed64)
	var scratch [8]byte
	binary.LittleEndian.PutUint64(scratch[:], math.Float64bits(value))
	return append(buf, scratch[:]...)
}

func readFloat32(data []byte) (float32, int, error) {
	v, n, err := readFixed32(data)
	if err != nil {
		return 0, 0, err
	}
	return math.Float32frombits(v), n, nil
}

func readFloat64(data []byte) (float64, int, error) {
	v, n, err := readFixed64(data)
	if err != nil {
		return 0, 0, err
	}
	return math.Float64frombits(v), n, nil
}

func readFixed32(data []byte) (uint32, int, error) {
	if len(data) < 4 {
		return 0, 0, errors.New("unexpected EOF reading fixed32")
	}
	return binary.LittleEndian.Uint32(data[:4]), 4, nil
}

func readFixed64(data []byte) (uint64, int, error) {
	if len(data) < 8 {
		return 0, 0, errors.New("unexpected EOF reading fixed64")
	}
	return binary.LittleEndian.Uint64(data[:8]), 8, nil
}
