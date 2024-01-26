package amf

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStream(t *testing.T) {
	var bs []byte
	buf := bytes.NewBuffer(bs)
	e := NewEncoder(buf)

	in := 1
	err := e.Encode(&in)
	if err != nil {
		t.Error(err)
		return
	}

	var out int
	d := NewDecoder(buf)
	err = d.Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}
