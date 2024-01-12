package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	header := &TypeHeader{}

	str := "hello,world!"
	strData := header.Marshal(TypeString, int32(len(str)))

	strType, strLen, err := header.Unmarshal(strData)
	require.Equal(t, nil, err)
	require.Equal(t, TypeString, strType)
	require.Equal(t, int32(len(str)), strLen)

	s := NewString()
	sData := s.Marshal(str)
	si, err := s.Unmarshal(sData)
	require.Equal(t, nil, err)
	require.Equal(t, str, si.(string))

	array := NewByteArrary()
	bytes := []byte("hello,world!")
	bytesData := array.Marshal(bytes)
	bytesi, err := array.Unmarshal(bytesData)
	require.Equal(t, nil, err)
	require.Equal(t, bytes, bytesi.([]byte))

	b := NewBool()
	bo := false
	bData := b.Marshal(bo)
	bi, err := b.Unmarshal(bData)
	require.Equal(t, nil, err)
	require.Equal(t, bo, bi.(bool))

	i32 := NewInt32()
	iv32 := int32(100)
	i32Data := i32.Marshal(iv32)
	i32i, err := i32.Unmarshal(i32Data)
	require.Equal(t, nil, err)
	require.Equal(t, iv32, i32i.(int32))

	i64 := NewInt64()
	iv64 := int64(100)
	i64Data := i64.Marshal(iv64)
	i64i, err := i64.Unmarshal(i64Data)
	require.Equal(t, nil, err)
	require.Equal(t, iv64, i64i.(int64))

	u64 := NewUint64()
	uv64 := uint64(100)
	u64Data := u64.Marshal(uv64)
	u64i, err := u64.Unmarshal(u64Data)
	require.Equal(t, nil, err)
	require.Equal(t, uv64, u64i.(uint64))
}
