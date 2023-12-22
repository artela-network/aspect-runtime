package runtimetypes

import (
	"reflect"

	"github.com/pkg/errors"
)

var (
	_ IType = (*ByteArray)(nil)
	_ IType = (*String)(nil)
)

// ByteArray implements IType
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
	ptr, err := ctx.Memory().Allocate(size)
	if err != nil {
		return 0, errors.Wrap(err, "allocate memory")
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
	ptr, err := ctx.Memory().Allocate(size)
	if err != nil {
		return 0, errors.Wrap(err, "allocate memory")
	}

	s.HStore(ctx, ptr)
	ctx.Memory().Write(ptr+s.HLen(), []byte(s.body))

	return ptr, nil
}

func (s *String) Load(ctx *Context, ptr int32) {
	// s.TypeHeader.HLoad(ctx, ptr)
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

func (b *Bool) Store(ctx *Context) (int32, error) {
	size := b.HLen() + b.dataLen
	ptr, err := ctx.Memory().Allocate(size)
	if err != nil {
		return 0, errors.Wrap(err, "allocate memory")
	}

	b.HStore(ctx, ptr)
	data := byte(0)
	if b.body {
		data = byte(1)
	}
	ctx.Memory().Write(ptr+b.HLen(), []byte{data})

	return ptr, nil
}

func (b *Bool) Load(ctx *Context, ptr int32) {
	b.TypeHeader.HLoad(ctx, ptr)
	b.body = false
	if ctx.Memory().Read(ptr+b.HLen(), b.dataLen)[0] == 1 {
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

type Int32 struct {
	TypeHeader

	body int32
}

func NewInt32() *Int32 {
	return &Int32{
		TypeHeader: TypeHeader{
			dataType: int16(TypeInt32),
		},
	}
}

func (i *Int32) Store(ctx *Context) (int32, error) {
	size := i.HLen() + i.dataLen
	ptr, err := ctx.Memory().Allocate(size)
	if err != nil {
		return 0, errors.Wrap(err, "allocate memory")
	}

	i.HStore(ctx, ptr)
	data := int32ToBytes(i.body)
	ctx.Memory().Write(ptr+i.HLen(), data)

	return ptr, nil
}

func (i *Int32) Load(ctx *Context, ptr int32) {
	i.TypeHeader.HLoad(ctx, ptr)
	data := ctx.Memory().Read(ptr+i.HLen(), i.dataLen)
	i.body = bytesToInt32(data)
}

func (i *Int32) Set(value interface{}) error {
	data, ok := value.(int32)
	if !ok {
		return errors.New("value is not bool")
	}
	i.dataLen = 4
	i.body = data
	return nil
}

func (i *Int32) Get() interface{} {
	return i.body
}

func (i *Int32) DataType() reflect.Type {
	return reflect.TypeOf(i.body)
}

type Int64 struct {
	TypeHeader

	body int64
}

func NewInt64() *Int64 {
	return &Int64{
		TypeHeader: TypeHeader{
			dataType: int16(TypeInt64),
		},
	}
}

func (i *Int64) Store(ctx *Context) (int32, error) {
	size := i.HLen() + i.dataLen
	ptr, err := ctx.Memory().Allocate(size)
	if err != nil {
		return 0, errors.Wrap(err, "allocate memory")
	}

	i.HStore(ctx, ptr)
	data := int64ToBytes(i.body)
	ctx.Memory().Write(ptr+i.HLen(), data)

	return ptr, nil
}

func (i *Int64) Load(ctx *Context, ptr int32) {
	i.TypeHeader.HLoad(ctx, ptr)
	data := ctx.Memory().Read(ptr+i.HLen(), i.dataLen)
	i.body = bytesToInt64(data)
}

func (i *Int64) Set(value interface{}) error {
	data, ok := value.(int64)
	if !ok {
		return errors.New("value is not bool")
	}
	i.dataLen = 8
	i.body = data
	return nil
}

func (i *Int64) Get() interface{} {
	return i.body
}

func (i *Int64) DataType() reflect.Type {
	return reflect.TypeOf(i.body)
}
