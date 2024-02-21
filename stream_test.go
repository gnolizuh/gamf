package amf

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestStream(t *testing.T) {
	var bs []byte
	buf := bytes.NewBuffer(bs)

	in := 1
	err := NewEncoder().WithWriter(buf).Encode(&in)
	if err != nil {
		t.Error(err)
		return
	}

	var out int
	err = NewDecoder().WithReader(buf).Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}
