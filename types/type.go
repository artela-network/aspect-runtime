package types

import "github.com/pkg/errors"

var typeNames = [12]string{
	"Unknown",
	"Int8",
	"Int16",
	"Int32",
	"Int64",
	"Uint8",
	"Uint16",
	"Uint32",
	"Uint64",
	"Bool",
	"String",
	"ByteArray",
}

// TypeIndex defines the index of runtime type
type TypeIndex int16

func (t TypeIndex) String() string {
	if t < 0 || int(t) >= len(typeNames) {
		return typeNames[0]
	}
	return typeNames[t]
}

const (
	TypeEmpty TypeIndex = iota // wrong type, no handler for empty type
	TypeInt8
	TypeInt16
	TypeInt32
	TypeInt64
	TypeUint8
	TypeUint16
	TypeUint32
	TypeUint64
	TypeBool
	TypeString // string with utf-8 encoder
	TypeByteArray
)

func AssertType(v interface{}) TypeIndex {
	switch v.(type) {
	case int8:
		return TypeInt8
	case int16:
		return TypeInt16
	case int32:
		return TypeInt32
	case int64:
		return TypeInt64
	case uint8:
		return TypeUint8
	case uint16:
		return TypeUint16
	case uint32:
		return TypeUint32
	case uint64:
		return TypeUint64
	case bool:
		return TypeBool
	case string:
		return TypeString
	case []byte:
		return TypeByteArray

		// for struct
		// case MyStruct, *MyStruct:
		// return TypeMyStruct
	}
	return TypeEmpty
}

func TypeObjectMapping(index TypeIndex) (IType, error) {
	switch index {
	case TypeByteArray:
		return NewByteArrary(), nil
	case TypeString:
		return NewString(), nil
	case TypeBool:
		return NewBool(), nil
	case TypeInt32:
		return NewInt32(), nil
	case TypeInt64:
		return NewInt64(), nil
	case TypeUint64:
		return NewUint64(), nil
	default:
		return nil, errors.Errorf("type of index %d is not valid", index)
	}
}
