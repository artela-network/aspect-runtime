package wasmtime

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrite(t *testing.T) {
	m := &Memory{
		data: make([]byte, 50),
		allocate: func(int32) (int32, error) {
			return 10, nil
		},
		refresh: func(m *Memory) {
		},
	}

	ptr, err := m.Write("hello")
	require.Equal(t, nil, err)
	require.Equal(t, int32(10), ptr)
	expectedHello := []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x000\x000\x000\x000\x000\x000\x000\x005\x00h\x00e\x00l\x00l\x00o\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	require.Equal(t, true, reflect.DeepEqual(expectedHello, m.data))

	m.data = make([]byte, 50)
	ptr, err = m.Write("")
	require.Equal(t, nil, err)
	require.Equal(t, int32(10), ptr)

	expectEmpty := []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x000\x000\x000\x000\x000\x000\x000\x000\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	require.Equal(t, true, reflect.DeepEqual(expectEmpty, m.data))
}

func TestRead(t *testing.T) {
	m := &Memory{
		data: []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x000\x000\x000\x000\x000\x000\x000\x005\x00h\x00e\x00l\x00l\x00o\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"),
		allocate: func(int32) (int32, error) {
			return 10, nil
		},
		refresh: func(m *Memory) {
		},
	}
	res, err := m.Read(10)
	require.Equal(t, nil, err)
	expectedHello := "hello"
	require.Equal(t, expectedHello, res)

	m.data = []byte("\x00\x00\x00\x00\x00\x00\x00\x00\x00\x000\x000\x000\x000\x000\x000\x000\x000\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	res, err = m.Read(10)
	require.Equal(t, nil, err)
	expectedEmptry := ""
	require.Equal(t, expectedEmptry, res)
}
