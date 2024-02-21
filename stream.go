//
// Copyright [2024] [https://github.com/gnolizuh]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package amf

import (
	"errors"
	"io"
)

// An Encoder writes AMF values to an output stream.
type Encoder struct {
	w   io.Writer
	v3  bool
	err error
}

// NewEncoder returns a new encoder.
func NewEncoder() *Encoder {
	return &Encoder{}
}

// WithWriter set Writer.
func (enc *Encoder) WithWriter(w io.Writer) *Encoder {
	enc.w = w
	return enc
}

// WithVersion set AMF codec version.
func (enc *Encoder) WithVersion(v Version) *Encoder {
	enc.v3 = v == Version3
	return enc
}

// Encode writes the AMF encoding of v to the stream,
// followed by a newline character.
//
// See the documentation for Marshal for details about the
// conversion of Go values to AMF.
func (enc *Encoder) Encode(v any) error {
	if enc.err != nil {
		return enc.err
	}

	if enc.w == nil {
		return errors.New("writer must be set")
	}

	e := newEncodeState()
	defer encodeStatePool.Put(e)

	err := e.marshal(v, encOpts{v3: enc.v3})
	if err != nil {
		return err
	}

	b := e.Bytes()
	if _, err = enc.w.Write(b); err != nil {
		enc.err = err
	}
	return err
}

// A Decoder reads and decodes AMF values from an input stream.
type Decoder struct {
	r   io.Reader
	v3  bool
	buf []byte
	d   decodeState
	err error
}

// NewDecoder returns a new decoder that reads from r.
//
// The decoder introduces its own buffering and may
// read data from r beyond the AMF values requested.
func NewDecoder() *Decoder {
	return &Decoder{}
}

// WithReader set Reader.
func (dec *Decoder) WithReader(r io.Reader) *Decoder {
	dec.r = r
	return dec
}

// WithVersion set AMF codec version.
func (dec *Decoder) WithVersion(v Version) *Decoder {
	dec.v3 = v == Version3
	return dec
}

// Decode reads the next AMF-encoded value from its
// input and stores it in the value pointed to by v.
//
// See the documentation for Unmarshal for details about
// the conversion of AMF into a Go value.
func (dec *Decoder) Decode(v any) error {
	if dec.err != nil {
		return dec.err
	}

	if dec.r == nil {
		return errors.New("reader must be set")
	}

	var err error
	dec.buf, err = io.ReadAll(dec.r)
	if err != nil {
		return err
	}
	dec.d.init(dec.buf, dec.v3)

	// Don't save err from unmarshal into dec.err:
	// the connection is still usable since we read a complete AMF
	// object from it before the error happened.
	return dec.d.unmarshal(v, decOpts{v3: dec.v3})
}
