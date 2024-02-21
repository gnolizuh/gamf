package amf

import (
	"bytes"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInt0(t *testing.T) {
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

func TestInt3(t *testing.T) {
	var bs []byte
	buf := bytes.NewBuffer(bs)

	in := 1
	err := NewEncoder().WithWriter(buf).WithVersion(Version3).Encode(&in)
	if err != nil {
		t.Error(err)
		return
	}

	var out int
	err = NewDecoder().WithReader(buf).WithVersion(Version3).Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestUInt0(t *testing.T) {
	in := uint(1)
	bs, err := Marshal(in)
	if err != nil {
		t.Error(err)
		return
	}

	var out uint
	err = Unmarshal(bs, &out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestUInt3(t *testing.T) {
	var bs []byte
	buf := bytes.NewBuffer(bs)

	in := uint(1)
	err := NewEncoder().WithWriter(buf).WithVersion(Version3).Encode(&in)
	if err != nil {
		t.Error(err)
		return
	}

	var out uint
	err = NewDecoder().WithReader(buf).WithVersion(Version3).Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestFloat0(t *testing.T) {
	in := 1.0
	bs, err := Marshal(in)
	if err != nil {
		t.Error(err)
		return
	}

	var out float64
	err = Unmarshal(bs, &out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestFloat3(t *testing.T) {
	var bs []byte
	buf := bytes.NewBuffer(bs)

	in := 1.0
	err := NewEncoder().WithWriter(buf).WithVersion(Version3).Encode(&in)
	if err != nil {
		t.Error(err)
		return
	}

	var out float64
	err = NewDecoder().WithReader(buf).WithVersion(Version3).Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestString0(t *testing.T) {
	in := "1"
	bs, err := Marshal(in)
	if err != nil {
		t.Error(err)
		return
	}

	var out string
	err = Unmarshal(bs, &out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestString3(t *testing.T) {
	var bs []byte
	buf := bytes.NewBuffer(bs)

	in := "1"
	err := NewEncoder().WithWriter(buf).WithVersion(Version3).Encode(&in)
	if err != nil {
		t.Error(err)
		return
	}

	var out string
	err = NewDecoder().WithReader(buf).WithVersion(Version3).Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestBool0(t *testing.T) {
	in := true
	bs, err := Marshal(&in)
	if err != nil {
		t.Error(err)
		return
	}

	out := false
	err = Unmarshal(bs, &out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestBool3(t *testing.T) {
	var bs []byte
	buf := bytes.NewBuffer(bs)

	in := true
	err := NewEncoder().WithWriter(buf).WithVersion(Version3).Encode(&in)
	if err != nil {
		t.Error(err)
		return
	}

	out := false
	err = NewDecoder().WithReader(buf).WithVersion(Version3).Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

// TestStruct1 make struct as input and output type.
func TestStruct2Struct(t *testing.T) {
	type Struct struct {
		Int    int    `amf:"tag_int"`
		String string `amf:"tag_string"`
		Bool   bool   `amf:"tag_bool"`
		Object struct {
			Int    int    `amf:"tag_int"`
			String string `amf:"tag_string"`
			Bool   bool   `amf:"tag_bool"`
		} `amf:"tag_object"`
	}

	in := Struct{Int: 1, String: "1", Bool: true}
	in.Object.Int = 1
	in.Object.String = "1"
	in.Object.Bool = true

	bs, err := Marshal(in)
	if err != nil {
		t.Error(err)
		return
	}

	out := Struct{}
	err = Unmarshal(bs, &out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

// TestStruct2 make struct as input type and map as output type.
func TestStruct2Map(t *testing.T) {
	type Struct struct {
		Int    int    `amf:"tag_int"`
		String string `amf:"tag_string"`
		Bool   bool   `amf:"tag_bool"`
		Object struct {
			Int    int    `amf:"tag_int"`
			String string `amf:"tag_string"`
			Bool   bool   `amf:"tag_bool"`
		} `amf:"tag_object"`
	}

	in := Struct{Int: 1, String: "1", Bool: true}
	in.Object.Int = 1
	in.Object.String = "1"
	in.Object.Bool = true

	bs, err := Marshal(in)
	if err != nil {
		t.Error(err)
		return
	}

	m := make(map[string]any)
	err = Unmarshal(bs, &m)
	if err != nil {
		t.Error(err)
		return
	}

	out := Struct{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "amf", Result: &out})
	if err != nil {
		t.Error(err)
		return
	}

	err = decoder.Decode(m)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

// TestStruct3 make map as input type and struct as output type.
func TestMap2Struct(t *testing.T) {
	type Struct struct {
		Int    int    `amf:"tag_int"`
		String string `amf:"tag_string"`
		Bool   bool   `amf:"tag_bool"`
		Object struct {
			Int    int    `amf:"tag_int"`
			String string `amf:"tag_string"`
			Bool   bool   `amf:"tag_bool"`
		} `amf:"tag_object"`
	}

	m := map[string]any{
		"tag_int":    1,
		"tag_string": "1",
		"tag_bool":   true,
		"tag_object": map[string]any{
			"tag_int":    1,
			"tag_string": "1",
			"tag_bool":   true,
		},
	}
	bs, err := Marshal(m)
	if err != nil {
		t.Error(err)
		return
	}

	out := Struct{}
	err = Unmarshal(bs, &out)
	if err != nil {
		t.Error(err)
		return
	}

	in := Struct{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "amf", Result: &in})
	if err != nil {
		t.Error(err)
		return
	}

	err = decoder.Decode(m)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestSlice0(t *testing.T) {
	in := []any{1.0, "1", true, map[string]any{"Int": 1.0, "String": "1", "Bool": true}}
	bs, err := Marshal(in)
	if err != nil {
		t.Error(err)
		return
	}

	out := []any{0.0, "0", false, map[string]any{}}
	err = Unmarshal(bs, &out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}

func TestSlice3(t *testing.T) {
	var bs []byte
	buf := bytes.NewBuffer(bs)

	in := []any{1.0, "1", true, map[string]any{"Int": 1.0, "String": "1", "Bool": true}}
	err := NewEncoder().WithWriter(buf).WithVersion(Version3).Encode(&in)
	if err != nil {
		t.Error(err)
		return
	}

	out := []any{0.0, "0", false, map[string]any{}}
	err = NewDecoder().WithReader(buf).WithVersion(Version3).Decode(&out)
	if err != nil {
		t.Error(err)
		return
	}

	require.Equal(t, in, out)
}
