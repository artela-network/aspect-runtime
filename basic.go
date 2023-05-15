package runtime

import (
	"reflect"

	"github.com/pkg/errors"
)

var (
	_ IType = (*ByteArray)(nil)
	_ IType = (*String)(nil)
)

// ByteArrary implements IType
type ByteArray struct {
	TypeHeader

	body []byte
}

func NewByteArrary() *ByteArray {
	return &ByteArray{
		TypeHeader: TypeHeader{
			dataType: int16(TypeByteArray),
		},
	}
}

func (b *ByteArray) Store(ctx *Context) (int32, error) {
	size := b.HLen() + b.dataLen
	ptr, err := ctx.Memory().alloc(size)
	if err != nil {
		return 0, errors.Wrap(err, "alloc memory")
	}

	b.HStore(ctx, ptr)
	ctx.Memory().Write(ptr+b.HLen(), b.body)

	return ptr, nil
}

func (b *ByteArray) Load(ctx *Context, ptr int32) {
	b.TypeHeader.HLoad(ctx, ptr)
	b.body = ctx.Memory().Read(ptr+b.HLen(), b.dataLen)
}

func (b *ByteArray) Set(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.New("value is not []byte")
	}
	b.dataLen = int32(len(data))
	b.body = data

	return nil
}

func (b *ByteArray) Get() interface{} {
	return b.body
}

func (b *ByteArray) DataType() reflect.Type {
	return reflect.TypeOf(b.body)
}

type String struct {
	TypeHeader

	body string
}

func NewString() *String {
	return &String{
		TypeHeader: TypeHeader{
			dataType: int16(TypeString),
		},
	}
}

func (s *String) Store(ctx *Context) (int32, error) {
	size := s.HLen() + s.dataLen
	ptr, err := ctx.Memory().alloc(size)
	if err != nil {
		return 0, errors.Wrap(err, "alloc memory")
	}

	s.HStore(ctx, ptr)
	ctx.Memory().Write(ptr+s.HLen(), []byte(s.body))

	return ptr, nil
}

func (s *String) Load(ctx *Context, ptr int32) {
	s.TypeHeader.HLoad(ctx, ptr)
	s.body = string(ctx.Memory().Read(ptr+s.HLen(), s.dataLen))
}

func (s *String) Set(value interface{}) error {
	data, ok := value.(string)
	if !ok {
		return errors.New("value is not string")
	}
	s.dataLen = int32(len(data))
	s.body = data
	return nil
}

func (s *String) Get() interface{} {
	return s.body
}

func (s *String) DataType() reflect.Type {
	return reflect.TypeOf(s.body)
}

type Bool struct {
	TypeHeader

	body bool
}

func NewBool() *Bool {
	return &Bool{
		TypeHeader: TypeHeader{
			dataType: int16(TypeBool),
		},
	}
}

func (b *Bool) Store() (int32, error) {
	size := b.HLen() + b.dataLen
	ptr, err := MemoryInstance().alloc(size)
	if err != nil {
		return 0, errors.Wrap(err, "alloc memory")
	}

	b.HStore(ptr)
	data := byte(0)
	if b.body {
		data = byte(1)
	}
	MemoryInstance().Write(ptr+b.HLen(), []byte{data})

	return ptr, nil
}

func (b *Bool) Load(ptr int32) {
	b.TypeHeader.HLoad(ptr)
	b.body = false
	if MemoryInstance().Read(ptr+b.HLen(), b.dataLen)[0] == 1 {
		b.body = true
	}
}

func (b *Bool) Set(value interface{}) error {
	data, ok := value.(bool)
	if !ok {
		return errors.New("value is not bool")
	}
	b.dataLen = 1
	b.body = data
	return nil
}

func (b *Bool) Get() interface{} {
	return b.body
}

func (b *Bool) DataType() reflect.Type {
	return reflect.TypeOf(b.body)
}
