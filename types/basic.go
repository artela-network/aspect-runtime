package types

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

const (
	HeaderDataType = 2
	HeaderDataLen  = 4
	HeaderLen      = 6
)

var (
	_ IType = (*ByteArray)(nil)
	_ IType = (*String)(nil)
	_ IType = (*Bool)(nil)
	_ IType = (*Int32)(nil)
	_ IType = (*Int64)(nil)
	_ IType = (*Uint64)(nil)
)

type TypeHeader struct{}

func (t *TypeHeader) Marshal(dataType TypeIndex, dataLen int32) []byte {
	var data [HeaderLen]byte
	data[0] = uint8(dataType)
	data[1] = uint8(dataType >> 8)

	dlen := int32ToBytes(dataLen)
	copy(data[2:], dlen)

	return data[:]
}

// Unmarshal desialize the data to the type
func (t *TypeHeader) Unmarshal(data []byte) (dataType TypeIndex, dataLen int32, err error) {
	if len(data) < HeaderLen {
		return 0, 0, errors.New("data is not valid, read header failed")
	}

	dataType = TypeIndex(int16(data[0]) + int16(data[1])<<8)
	dataLen = bytesToInt32(data[2:6])
	return dataType, dataLen, nil
}

// ByteArray implements IType
type ByteArray struct {
	dataType TypeIndex
	dataLen  int32
}

func NewByteArrary() *ByteArray {
	return &ByteArray{dataType: TypeByteArray}
}

func (t *ByteArray) Marshal(value interface{}) []byte {
	input, ok := value.([]byte)
	if !ok {
		panic(fmt.Sprintf("expected a []byte, but the input was %d", reflect.TypeOf(value)))
	}
	t.dataLen = int32(len(input))

	data := make([]byte, t.dataLen+HeaderLen)
	header := &TypeHeader{}
	headerData := header.Marshal(t.dataType, t.dataLen)
	copy(data[:], headerData)

	copy(data[HeaderLen:], input)

	return data
}

// Unmarshal desialize the data to the type
func (t *ByteArray) Unmarshal(data []byte) (interface{}, error) {
	header := TypeHeader{}
	_, dataLen, err := header.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	return data[HeaderLen : HeaderLen+dataLen], nil
}

// String implements IType
type String struct {
	dataType TypeIndex
	dataLen  int32
}

func NewString() *String {
	return &String{dataType: TypeString}
}

func (t *String) Marshal(value interface{}) []byte {
	input, ok := value.(string)
	if !ok {
		panic(fmt.Sprintf("expected a string, but the input was %d", reflect.TypeOf(value)))
	}
	t.dataLen = int32(len(input))

	data := make([]byte, t.dataLen+HeaderLen)
	header := &TypeHeader{}
	headerData := header.Marshal(t.dataType, t.dataLen)
	copy(data[:], headerData)

	copy(data[HeaderLen:], []byte(input))

	return data
}

// Unmarshal desialize the data to the type
func (t *String) Unmarshal(data []byte) (interface{}, error) {
	header := TypeHeader{}
	_, dataLen, err := header.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	return string(data[HeaderLen : HeaderLen+dataLen]), nil
}

type Bool struct {
	dataType TypeIndex
	dataLen  int32
}

func NewBool() *Bool {
	return &Bool{dataType: TypeBool}
}

func (t *Bool) Marshal(value interface{}) []byte {
	input, ok := value.(bool)
	if !ok {
		panic(fmt.Sprintf("expected a bool, but the input was %d", reflect.TypeOf(value)))
	}
	t.dataLen = 1

	data := make([]byte, t.dataLen+HeaderLen)
	header := &TypeHeader{}
	headerData := header.Marshal(t.dataType, t.dataLen)
	copy(data[:], headerData)

	body := byte(0)
	if input {
		body = byte(1)
	}
	data[HeaderLen] = body

	return data
}

// Unmarshal desialize the data to the type
func (t *Bool) Unmarshal(data []byte) (interface{}, error) {
	header := TypeHeader{}
	_, dataLen, err := header.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	return data[HeaderLen : HeaderLen+dataLen][0] == 1, nil
}

type Int32 struct {
	dataType TypeIndex
	dataLen  int32
}

func NewInt32() *Int32 {
	return &Int32{dataType: TypeInt32}
}

func (t *Int32) Marshal(value interface{}) []byte {
	input, ok := value.(int32)
	if !ok {
		panic(fmt.Sprintf("expected a int32, but the input was %d", reflect.TypeOf(value)))
	}
	t.dataLen = 4

	data := make([]byte, t.dataLen+HeaderLen)
	header := &TypeHeader{}
	headerData := header.Marshal(t.dataType, t.dataLen)
	copy(data[:], headerData)

	body := int32ToBytes(input)
	copy(data[HeaderLen:], body)

	return data
}

// Unmarshal desialize the data to the type
func (t *Int32) Unmarshal(data []byte) (interface{}, error) {
	header := TypeHeader{}
	_, dataLen, err := header.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	body := data[HeaderLen : HeaderLen+dataLen]
	return bytesToInt32(body), nil
}

type Int64 struct {
	dataType TypeIndex
	dataLen  int32
}

func NewInt64() *Int64 {
	return &Int64{dataType: TypeInt64}
}

func (t *Int64) Marshal(value interface{}) []byte {
	input, ok := value.(int64)
	if !ok {
		panic(fmt.Sprintf("expected a int64, but the input was %d", reflect.TypeOf(value)))
	}
	t.dataLen = 8

	data := make([]byte, t.dataLen+HeaderLen)
	header := &TypeHeader{}
	headerData := header.Marshal(t.dataType, t.dataLen)
	copy(data[:], headerData)

	body := int64ToBytes(input)
	copy(data[HeaderLen:], body)

	return data
}

// Unmarshal desialize the data to the type
func (t *Int64) Unmarshal(data []byte) (interface{}, error) {
	header := TypeHeader{}
	_, dataLen, err := header.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	body := data[HeaderLen : HeaderLen+dataLen]
	return bytesToInt64(body), nil
}

type Uint64 struct {
	dataType TypeIndex
	dataLen  int32
}

func NewUint64() *Uint64 {
	return &Uint64{dataType: TypeUint64}
}

func (t *Uint64) Marshal(value interface{}) []byte {
	input, ok := value.(uint64)
	if !ok {
		panic(fmt.Sprintf("expected a int64, but the input was %d", reflect.TypeOf(value)))
	}
	t.dataLen = 8

	data := make([]byte, t.dataLen+HeaderLen)
	header := &TypeHeader{}
	headerData := header.Marshal(t.dataType, t.dataLen)
	copy(data[:], headerData)

	body := int64ToBytes(int64(input))
	copy(data[HeaderLen:], body)

	return data
}

// Unmarshal desialize the data to the type
func (t *Uint64) Unmarshal(data []byte) (interface{}, error) {
	header := TypeHeader{}
	_, dataLen, err := header.Unmarshal(data)
	if err != nil {
		return nil, err
	}

	body := data[HeaderLen : HeaderLen+dataLen]
	return uint64(bytesToInt64(body)), nil
}
