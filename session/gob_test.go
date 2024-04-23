package session

import (
	"bytes"
	"encoding/gob"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

type TestInterface interface {
	Hello() string
}

type Ttt struct {
	Name string
}

func (t *Ttt) Hello() string {
	return "hello"
}

func NewTest() TestInterface {
	return &Ttt{
		Name: "test",
	}
}

func TestGob(t *testing.T) {
	gob.Register(Ttt{})
	var a TestInterface = &Ttt{"test"}
	encodedBuf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(encodedBuf).Encode(a)
	require.NoError(t, err)
	b := NewTest()
	typeB := reflect.TypeOf(b)
	valB := reflect.New(typeB)

	err = gob.NewDecoder(encodedBuf).DecodeValue(valB)
	require.NoError(t, err)
	require.Equal(t, a.Hello(), b.Hello())
}

func TestGobSession(t *testing.T) {
	var a Session = &session{
		Id: "test",
		Data: map[string]any{
			"test":  "test",
			"test2": 1,
		},
	}

	encodedBuf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(encodedBuf).Encode(a)
	require.NoError(t, err)

	b := DefaultBuilder(nil, "")
	typeB := reflect.TypeOf(b)
	valB := reflect.New(typeB).Elem()
	decodedBuf := bytes.NewBuffer(encodedBuf.Bytes())
	err = gob.NewDecoder(decodedBuf).DecodeValue(valB)
	require.NoError(t, err)

	b = valB.Interface().(Session)
	id := b.ID()
	require.Equal(t, a.ID(), id)

	val, err := b.Get("test")
	require.NoError(t, err)
	require.Equal(t, "test", val)

}
