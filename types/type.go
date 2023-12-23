package runtimetypes

// TypeIndex defines the index of runtime type
type TypeIndex int16

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

var TypeObjectMapping = map[TypeIndex]IType{
	TypeByteArray: NewByteArrary(),
	TypeString:    NewString(),
	TypeBool:      NewBool(),
	TypeInt32:     NewInt32(),
	TypeInt64:     NewInt64(),
	TypeUint64:    NewUInt64(),
}

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
