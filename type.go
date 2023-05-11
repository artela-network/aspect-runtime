package runtime

// TypeIndex defines the index of runtime type
type TypeIndex int16

const (
	// TypeEmpty
	TypeInt8 TypeIndex = iota
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
}

var TypeMapping = map[string]TypeIndex{
	"[]byte":  TypeByteArray,
	"string":  TypeString,
	"*string": TypeString,
}
