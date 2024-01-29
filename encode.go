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
	"bytes"
	"encoding/binary"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

type EncoderError struct {
	Type reflect.Type
	Err  error
}

func (e *EncoderError) Error() string {
	return "amf: error calling EncodeAMF for type " + e.Type.String() + ": " + e.Err.Error()
}

// A MarshalerError represents an error from calling a MarshalAMF method.
type MarshalerError struct {
	Type       reflect.Type
	Err        error
	sourceFunc string
}

func (e *MarshalerError) Error() string {
	srcFunc := e.sourceFunc
	if srcFunc == StringEmpty {
		srcFunc = "MarshalAMF"
	}
	return "amf: error calling " + srcFunc +
		" for type " + e.Type.String() +
		": " + e.Err.Error()
}

// An encodeState encodes AMF into a bytes.Buffer.
type encodeState struct {
	bytes.Buffer // accumulated output

	strReference Reference
	objReference Reference
}

var encodeStatePool sync.Pool

func newEncodeState() *encodeState {
	if v := encodeStatePool.Get(); v != nil {
		e := v.(*encodeState)
		e.Reset()
		return e
	}
	return &encodeState{
		strReference: Reference{m: make(map[interface{}]int), a: make([]interface{}, 0)},
		objReference: Reference{m: make(map[interface{}]int), a: make([]interface{}, 0)},
	}
}

// amfError is an error wrapper type for internal use only.
// Panics with errors are wrapped in amfError so that the top-level recover
// can distinguish intentional panics from this package.
type amfError struct{ error }

func (e *encodeState) marshal(v interface{}, opts encOpts) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if s, ok := r.(string); ok {
				panic(s)
			}
			err = r.(error)
		}
	}()
	e.reflectValue(reflect.ValueOf(v), opts)
	return nil
}

// error aborts the encoding by panicking with err wrapped in amfError.
func (e *encodeState) error(err error) {
	panic(amfError{err})
}

func (e *encodeState) encodeError(err error) {
	panic(err)
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return v.Bool() == false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Pointer:
		return v.IsNil()
	}
	return false
}

func (e *encodeState) reflectValue(v reflect.Value, opts encOpts) {
	valueEncoder(v)(e, v, opts)
}

type encOpts struct {
	// ver3 means AMF version, which is true for AMF3, default is AMF0.
	ver3 bool

	// quoted causes primitive fields to be encoded inside AMF strings.
	quoted bool
}

func (e *encodeState) encodeObjectName(s string) {
	e.writeUint(uint16(len(s)))
	e.Write([]byte(s))
}

func (e *encodeState) writeMarker(m byte) {
	e.WriteByte(m)
}

type reflectWithString struct {
	v reflect.Value
	s string
}

func (w *reflectWithString) resolve() error {
	if w.v.Kind() == reflect.String {
		w.s = w.v.String()
		return nil
	}
	switch w.v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		w.s = strconv.FormatInt(w.v.Int(), 10)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		w.s = strconv.FormatUint(w.v.Uint(), 10)
		return nil
	}
	panic("unexpected map key type")
}

// encodeString encoding rule:
// AMF0:
//   - len(s) <= 0: string-marker[1B] + length[2B] + s
//   - len(s) > 0xffff: long-string-marker[1B] + length[4B] + s
//
// AMF3:
//   - no cache: string-marker[1B] + length[29b] + s
//   - cached: string-marker[1B] + index[29b]
func (e *encodeState) encodeString(s string, opts encOpts) {
	if !opts.ver3 {
		l := len(s)
		if l > LongStringSize {
			e.writeMarker(LongStringMarker0)
			e.writeUint(uint32(l))
		} else {
			e.writeMarker(StringMarker0)
			e.writeUint(uint16(l))
		}
		e.Write([]byte(s))
	} else {
		e.writeMarker(StringMarker3)
		i, ok := e.strReference.get(s)
		if ok {
			e.writeU29(uint32(i << 1))
			return
		}
		l := len(s)
		e.writeU29(uint32((l << 1) | 1))
		if l > 0 {
			_ = e.strReference.set(s)
		}
		e.Write([]byte(s))
	}
}

func (e *encodeState) writeUint(u interface{}) {
	_ = binary.Write(e, binary.BigEndian, u)
}

// writeU29 AMF3 only.
func (e *encodeState) writeU29(v uint32) {
	v = v & 0x1fffffff
	b := make([]byte, 0, 4)
	switch {
	case v < 0x80:
		b = append(b, byte(v))
	case v < 0x4000:
		b = append(b, byte((v>>7)|0x80))
		b = append(b, byte(v&0x7f))
	case v < 0x200000:
		b = append(b, byte((v>>14)|0x80))
		b = append(b, byte((v>>7)|0x80))
		b = append(b, byte(v&0x7f))
	case v < 0x20000000:
		b = append(b, byte((v>>22)|0x80))
		b = append(b, byte((v>>15)|0x80))
		b = append(b, byte((v>>7)|0x80))
		b = append(b, byte(v&0xff))
	default:
		return
	}
	e.Write(b)
}

// encodeBool encoding rule:
//
// AMF0:
//   - bool-marker[1B] + b
//
// AMF3:
//   - true: true-marker[1B]
//   - false: false-marker[1B]
func (e *encodeState) encodeBool(b bool, opts encOpts) {
	if !opts.ver3 {
		e.writeMarker(BooleanMarker0)
		if b {
			e.WriteByte(1)
		} else {
			e.WriteByte(0)
		}
	} else {
		if b {
			e.WriteByte(TrueMarker3)
		} else {
			e.WriteByte(FalseMarker3)
		}
	}
}

// encodeInt encoding rule:
//
// AMF0:
//   - number-marker[1B] + i[8B]
//
// AMF3:
//   - -2^28 <= i <= 2^28: integer-marker[1B] + i[29b]
//   - else: encode as AMF3 double.
func (e *encodeState) encodeInt(i int64, opts encOpts) {
	if !opts.ver3 {
		e.encodeFloat64(float64(i), opts)
	} else {
		if i < Int28Min && i >= Int28Max {
			e.encodeFloat64(float64(i), opts)
		} else {
			e.writeMarker(IntegerMarker3)
			e.writeU29(uint32(i))
		}
	}
}

// encodeUInt it's same with encodeInt.
func (e *encodeState) encodeUInt(ui uint64, opts encOpts) {
	if !opts.ver3 {
		e.encodeFloat64(float64(ui), opts)
	} else {
		if ui >= UInt29Max {
			e.encodeFloat64(float64(ui), opts)
		} else {
			e.writeMarker(IntegerMarker3)
			e.writeU29(uint32(ui))
		}
	}
}

// encodeDate:
//
// - AMF0: date-marker[1B] DOUBLE[8B] time-zone[2B]
// - AMF3: date-marker (U29O-ref | (U29D-value date-time))
func (e *encodeState) encodeDate(v reflect.Value, opts encOpts) {
	t := v.Interface().(time.Time)
	if !opts.ver3 {
		e.writeMarker(DateMarker0)
		_ = binary.Write(e, binary.BigEndian, t.UnixMilli())
		_ = binary.Write(e, binary.BigEndian, 0x0000) // time-zone
	} else {
		e.writeMarker(DateMarker3)
		i, ok := e.strReference.get(v)
		if ok {
			e.writeU29(uint32(i << 1))
			return
		}
		_ = e.strReference.set(v)
		e.writeU29(uint32(0x01))
		e.writeUint(t.UnixMilli())
	}
}

// encodeFloat64 encoding rule:
//
// AMF0:
//   - number-marker[1B] + i[8B]
//
// AMF3:
//   - double-marker[1B] + i[8B]
func (e *encodeState) encodeFloat64(f float64, opts encOpts) {
	if !opts.ver3 {
		e.writeMarker(NumberMarker0)
		_ = binary.Write(e, binary.BigEndian, f)
	} else {
		e.writeMarker(DoubleMarker3)
		_ = binary.Write(e, binary.BigEndian, f)
	}
}

func (e *encodeState) encodeNull(opts encOpts) {
	if !opts.ver3 {
		e.WriteByte(NullMarker0)
	} else {
		e.WriteByte(NullMarker3)
	}
}

type encoderFunc func(e *encodeState, v reflect.Value, opts encOpts)

var encoderCache sync.Map

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}
	return typeEncoder(v.Type())
}

func typeEncoder(t reflect.Type) encoderFunc {
	if fi, ok := encoderCache.Load(t); ok {
		return fi.(encoderFunc)
	}

	var (
		wg sync.WaitGroup
		f  encoderFunc
	)
	wg.Add(1)
	fi, loaded := encoderCache.LoadOrStore(t, encoderFunc(func(e *encodeState, v reflect.Value, opts encOpts) {
		wg.Wait()
		f(e, v, opts)
	}))
	if loaded {
		return fi.(encoderFunc)
	}

	f = newTypeEncoder(t)
	wg.Done()
	encoderCache.Store(t, f)
	return f
}

var (
	marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()
)

func newTypeEncoder(t reflect.Type) encoderFunc {
	if t.Implements(marshalerType) {
		return marshalerEncoder
	}

	switch t.Kind() {
	case reflect.String:
		return stringEncoder
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintEncoder
	case reflect.Float32, reflect.Float64:
		return floatEncoder
	case reflect.Interface:
		return interfaceEncoder
	case reflect.Struct: // struct & map are type of Object Marker.
		return newStructEncoder(t)
	case reflect.Map:
		return newMapEncoder(t)
	case reflect.Slice: // slice & array are type of Strict Array Marker.
		return newSliceEncoder(t)
	case reflect.Array:
		return newArrayEncoder(t)
	case reflect.Ptr:
		return newPtrEncoder(t)
	default:
		return newDefaultEncoder()
	}
}

func marshalerEncoder(e *encodeState, v reflect.Value, _ encOpts) {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		e.WriteString("null")
		return
	}
	m, ok := v.Interface().(Marshaler)
	if !ok {
		e.WriteString("null")
		return
	}
	b, err := m.MarshalAMF()
	if err == nil {
		e.Grow(len(b))
		out := e.AvailableBuffer()
		e.Buffer.Write(out)
	}
	if err != nil {
		e.error(&MarshalerError{v.Type(), err, "MarshalAMF"})
	}
}

func stringEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	e.encodeString(v.String(), opts)
}

func boolEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	e.encodeBool(v.Bool(), opts)
}

func invalidValueEncoder(e *encodeState, _ reflect.Value, opts encOpts) {
	e.encodeNull(opts)
}

func intEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	e.encodeInt(v.Int(), opts)
}

func uintEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	e.encodeUInt(v.Uint(), opts)
}

func floatEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	e.encodeFloat64(v.Float(), opts)
}

func dateEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	e.encodeDate(v, opts)
}

func interfaceEncoder(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		e.encodeNull(opts)
		return
	}
	e.reflectValue(v.Elem(), opts)
}

func isValidTag(s string) bool {
	if s == StringEmpty {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
}

type structEncoder struct {
	fields structFields
}

type structFields struct {
	list         []field
	byExactName  map[string]*field
	byFoldedName map[string]*field
}

// structEncoder encoding rule:
//
// AMF0:
//   - object-marker[1B] + [object...] + "" + end-object-marker[1B]
//
// AMF3:
//  1. object-marker U29O-ref *value-type *dynamic-member
//  2. object-marker U29O-traits-ext class-name *(UTF-8-vr) *value-type *dynamic-member
//  3. object-marker U29O-traits-ref *value-type *dynamic-member
//  4. object-marker U29O-traits class-name *(UTF-8-vr) *value-type *dynamic-member
func (se *structEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	// write start marker
	if !opts.ver3 {
		e.writeMarker(ObjectMarker0)
	} else {
		e.writeMarker(ObjectMarker3)
		i, ok := e.strReference.get(v)
		if ok {
			e.writeU29(uint32(i << 1))
			return
		}
		_ = e.strReference.set(v)
		e.writeMarker(0x0b)
		e.encodeString(StringEmpty, opts)
	}

FieldLoop:
	// write object
	for i := range se.fields.list {
		f := &se.fields.list[i]

		// Find the nested struct field by following f.index.
		fv := v
		for _, i := range f.index {
			if fv.Kind() == reflect.Pointer {
				if fv.IsNil() {
					continue FieldLoop
				}
				fv = fv.Elem()
			}
			fv = fv.Field(i)
		}

		if f.omitEmpty && isEmptyValue(fv) {
			continue
		}
		opts.quoted = f.quoted
		if !opts.ver3 {
			e.encodeObjectName(f.name)
		} else {
			e.encodeString(f.name, opts)
		}
		if f.xml && fv.Kind() == reflect.String {
			encodeXML(e, fv, opts)
		} else {
			f.encoder(e, fv, opts)
		}
	}

	// write end marker
	if !opts.ver3 {
		e.encodeObjectName(StringEmpty)
		e.writeMarker(ObjectEndMarker0)
	} else {
		e.encodeString(StringEmpty, opts)
	}
}

func newStructEncoder(t reflect.Type) encoderFunc {
	se := structEncoder{fields: cachedTypeFields(t)}
	return se.encode
}

type mapEncoder struct {
	elemEnc encoderFunc
}

func (mae *mapEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		e.encodeNull(opts)
		return
	}

	keys := v.MapKeys()
	sv := make([]reflectWithString, len(keys))
	for i, v := range keys {
		sv[i].v = v
		if err := sv[i].resolve(); err != nil {
			e.encodeError(&EncoderError{v.Type(), err})
		}
	}
	sort.Slice(sv, func(i, j int) bool { return sv[i].s < sv[j].s })

	// write start marker
	if !opts.ver3 {
		e.writeMarker(ObjectMarker0)
	} else {
		e.writeMarker(ObjectMarker3)
		i, ok := e.strReference.get(v)
		if ok {
			e.writeU29(uint32(i << 1))
			return
		}
		_ = e.strReference.set(v)
		e.writeMarker(0x0b)
		e.encodeString(StringEmpty, opts)
	}

	// write object
	for _, kv := range sv {
		if !opts.ver3 {
			e.encodeObjectName(kv.s)
		} else {
			e.encodeString(kv.s, opts)
		}
		mae.elemEnc(e, v.MapIndex(kv.v), opts)
	}

	// write end marker
	if !opts.ver3 {
		e.encodeObjectName(StringEmpty)
		e.writeMarker(ObjectEndMarker0)
	} else {
		e.encodeString(StringEmpty, opts)
	}
}

func newMapEncoder(t reflect.Type) encoderFunc {
	switch t.Key().Kind() {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
	default:
		return unsupportedTypeEncoder
	}
	mae := &mapEncoder{elemEnc: typeEncoder(t.Elem())}
	return mae.encode
}

func encodeXML(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		e.encodeNull(opts)
		return
	}

	s := v.Interface().(string)
	if !opts.ver3 {
		e.writeMarker(XMLDocumentMarker0)
		e.writeUint(uint32(len(s)))
		e.WriteString(s)
	} else {
		e.writeMarker(XMLDocMarker3)
		i, ok := e.strReference.get(v)
		if ok {
			e.writeU29(uint32(i << 1))
			return
		}
		_ = e.strReference.set(v)
		e.writeU29(uint32((i << 1) | 1))
		e.WriteString(s)
	}
}

type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se *sliceEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		e.encodeNull(opts)
		return
	}
	se.arrayEnc(e, v, opts)
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	enc := &sliceEncoder{arrayEnc: newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

// arrayEncoder
//   - AMF0: strict-array-marker array-count[4B] *(value-type)
//   - AMF3: array-marker (U29O-ref[29b] | (U29A-value (UTF-8-empty | *(assoc-value) UTF-8-empty) *(value-type)))
//     assoc-value means value with name, value-type means value with no name.
func (ae *arrayEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() || v.Len() == 0 {
		if !opts.ver3 {
			e.writeMarker(UndefinedMarker0)
		} else {
			e.writeMarker(UndefinedMarker3)
		}
		return
	}

	if !opts.ver3 {
		e.writeMarker(StrictArrayMarker0)
		e.writeUint(uint32(v.Len()))
	} else {
		e.writeMarker(ArrayMarker3)
		i, ok := e.strReference.get(v)
		if ok {
			e.writeU29(uint32(i << 1))
			return
		}
		_ = e.strReference.set(v)
		e.writeU29((uint32(v.Len()) << 1) | 1)
		e.WriteByte(UTF8Empty)
	}

	for i := 0; i < v.Len(); i++ {
		ae.elemEnc(e, v.Index(i), opts)
	}
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := &arrayEncoder{elemEnc: typeEncoder(t.Elem())}
	return enc.encode
}

type ptrEncoder struct {
	elemEnc encoderFunc
}

func (pe *ptrEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	if v.IsNil() {
		e.encodeNull(opts)
		return
	}
	pe.elemEnc(e, v.Elem(), opts)
}

func newPtrEncoder(t reflect.Type) encoderFunc {
	enc := &ptrEncoder{elemEnc: typeEncoder(t.Elem())}
	return enc.encode
}

type defaultEncoder struct {
	elemEnc encoderFunc
}

func newDefaultEncoder() encoderFunc {
	de := defaultEncoder{}
	return de.encode
}

func (de *defaultEncoder) encode(e *encodeState, v reflect.Value, opts encOpts) {
	switch v.Interface().(type) {
	case time.Time:
		dateEncoder(e, v, opts)
	default:
		unsupportedTypeEncoder(e, v, opts)
	}
}

func unsupportedTypeEncoder(e *encodeState, v reflect.Value, _ encOpts) {
	e.encodeError(&UnsupportedTypeError{v.Type()})
}

type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "amf: unsupported type: " + e.Type.String()
}

// Marshal returns the AMF encoding of v.
//
// Marshal traverses the value v recursively.
// If an encountered value implements the Marshaler interface
// and is not a nil pointer, Marshal calls its MarshalAMF method
// to produce AMF.
//
// Otherwise, Marshal uses the following type-dependent default encodings:
//
// Boolean values encode as AMF booleans.
//
// Floating point, integer, and Number values encode as AMF numbers.
//
// String values encode as AMF strings coerced to valid UTF-8,
// replacing invalid bytes with the Unicode replacement rune.
//
// Array and slice values encode as AMF arrays, except that
// []byte encodes as a base64-encoded string, and a nil slice
// encodes as the null AMF value.
//
// Struct, Map values encode as AMF objects.
// Each exported struct field becomes a member of the object.
//
// Pointer values encode as the value pointed to.
// A nil pointer encodes as the null AMF value.
//
// Interface values encode as the value contained in the interface.
// A nil interface value encodes as the null AMF value.
//
// Channel, complex, and function values cannot be encoded in AMF.
// Attempting to encode such a value causes Marshal to return
// an UnsupportedTypeError.
//
// AMF cannot represent cyclic data structures and Marshal does not
// handle them. Passing cyclic structures to Marshal will result in
// an error.
func Marshal(v any) ([]byte, error) {
	e := newEncodeState()
	defer encodeStatePool.Put(e)

	err := e.marshal(v, encOpts{})
	if err != nil {
		return nil, err
	}
	b := append([]byte(nil), e.Bytes()...)

	return b, nil
}

// Marshaler is the interface implemented by types that
// can marshal themselves into valid AMF.
type Marshaler interface {
	MarshalAMF() ([]byte, error)
}

// A field represents a single field found in a struct.
type field struct {
	name      string
	nameBytes []byte // []byte(name)

	tag       bool
	index     []int
	typ       reflect.Type
	omitEmpty bool
	xml       bool
	quoted    bool

	encoder encoderFunc
}

// byIndex sorts field by index sequence.
type byIndex []field

func (x byIndex) Len() int { return len(x) }

func (x byIndex) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

func (x byIndex) Less(i, j int) bool {
	for k, xik := range x[i].index {
		if k >= len(x[j].index) {
			return false
		}
		if xik != x[j].index[k] {
			return xik < x[j].index[k]
		}
	}
	return len(x[i].index) < len(x[j].index)
}

// typeFields returns a list of fields that AMF should recognize for the given type.
// The algorithm is breadth-first search over the set of structs to include - the top struct
// and then any reachable anonymous structs.
func typeFields(t reflect.Type) structFields {
	// Anonymous fields to explore at the current level and the next.
	var current []field
	next := []field{{typ: t}}

	// Count of queued names for current level and the next.
	var count, nextCount map[reflect.Type]int

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var fields []field

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Pointer {
						t = t.Elem()
					}
					if !sf.IsExported() && t.Kind() != reflect.Struct {
						// Ignore embedded fields of unexported non-struct types.
						continue
					}
					// Do not ignore embedded fields of unexported struct types
					// since they may have exported fields.
				} else if !sf.IsExported() {
					// Ignore unexported non-embedded fields.
					continue
				}
				tag := sf.Tag.Get("amf")
				if tag == "-" {
					continue
				}
				name, opts := parseTag(tag)
				if !isValidTag(name) {
					name = StringEmpty
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Name() == StringEmpty && ft.Kind() == reflect.Pointer {
					// Follow pointer.
					ft = ft.Elem()
				}

				// Only strings, floats, integers, and booleans can be quoted.
				quoted := false
				if opts.Contains("string") {
					switch ft.Kind() {
					case reflect.Bool,
						reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
						reflect.Float32, reflect.Float64,
						reflect.String:
						quoted = true
					}
				}

				// Record found field and index sequence.
				if name != StringEmpty || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != StringEmpty
					if name == StringEmpty {
						name = sf.Name
					}
					field := field{
						name:      name,
						tag:       tagged,
						index:     index,
						typ:       ft,
						omitEmpty: opts.Contains("omitempty"),
						xml:       opts.Contains("xml"),
						quoted:    quoted,
					}
					field.nameBytes = []byte(field.name)

					fields = append(fields, field)
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 or 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, field{name: ft.Name(), index: index, typ: ft})
				}
			}
		}
	}

	sort.Slice(fields, func(i, j int) bool {
		x := fields
		// sort field by name, breaking ties with depth, then
		// breaking ties with "name came from amf tag", then
		// breaking ties with index sequence.
		if x[i].name != x[j].name {
			return x[i].name < x[j].name
		}
		if len(x[i].index) != len(x[j].index) {
			return len(x[i].index) < len(x[j].index)
		}
		if x[i].tag != x[j].tag {
			return x[i].tag
		}
		return byIndex(x).Less(i, j)
	})

	// Delete all fields that are hidden by the Go rules for embedded fields,
	// except that fields with AMF tags are promoted.

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	fields = out
	sort.Sort(byIndex(fields))

	for i := range fields {
		f := &fields[i]
		f.encoder = typeEncoder(typeByIndex(t, f.index))
	}
	exactNameIndex := make(map[string]*field, len(fields))
	foldedNameIndex := make(map[string]*field, len(fields))
	for i, field := range fields {
		exactNameIndex[field.name] = &fields[i]
		// For historical reasons, first folded match takes precedence.
		if _, ok := foldedNameIndex[string(foldName(field.nameBytes))]; !ok {
			foldedNameIndex[string(foldName(field.nameBytes))] = &fields[i]
		}
	}
	return structFields{fields, exactNameIndex, foldedNameIndex}
}

func dominantField(fields []field) (field, bool) {
	length := len(fields[0].index)
	tagged := -1 // Index of first tagged field.
	for i, f := range fields {
		if len(f.index) > length {
			fields = fields[:i]
			break
		}
		if f.tag {
			if tagged >= 0 {
				return field{}, false
			}
			tagged = i
		}
	}

	if tagged >= 0 {
		return fields[tagged], true
	}

	if len(fields) > 1 {
		return field{}, false
	}
	return fields[0], true
}

var fieldCache sync.Map // map[reflect.Type]structFields

// cachedTypeFields is like typeFields but uses a cache to avoid repeated work.
func cachedTypeFields(t reflect.Type) structFields {
	if f, ok := fieldCache.Load(t); ok {
		return f.(structFields)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.(structFields)
}
