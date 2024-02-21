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

type Version uint8

const (
	Version0 = Version(0)
	Version3 = Version(3)
)

// https://rtmp.veriskope.com/pdf/amf0-file-format-specification.pdf
const (
	NumberMarker0 = iota
	BooleanMarker0
	StringMarker0
	ObjectMarker0
	MovieClipMarker0
	NullMarker0
	UndefinedMarker0
	ReferenceMarker0
	ECMAArrayMarker0
	ObjectEndMarker0
	StrictArrayMarker0
	DateMarker0
	LongStringMarker0
	UnsupportedMarker0
	RecordSetMarker0
	XMLDocumentMarker0
	TypedObjectMarker0
	ACMPlusObjectMarker0
)

const (
	UndefinedMarker3 = iota
	NullMarker3
	FalseMarker3
	TrueMarker3
	IntegerMarker3
	DoubleMarker3
	StringMarker3
	XMLDocMarker3
	DateMarker3
	ArrayMarker3
	ObjectMarker3
	XMLMarker3
	ByteArrayMarker3
	VectorIntMarker3
	VectorUIntMarker3
	VectorDoubleMarker3
	VectorObjectMarker3
	DictionaryMarker3
)

const (
	// LongStringSize AMF0 only.
	LongStringSize = 0xffff

	UInt29Max = 0x1fffffff
	Int28Max  = 0xfffffff
	Int28Min  = -0xfffffff

	UTF8Empty   = 0x01
	U29NoTraits = 0x0b
	StringEmpty = ""
)
