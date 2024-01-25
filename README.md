# gamf

![Go](https://github.com/goccy/go-json/workflows/Go/badge.svg)
[![GoDoc](https://godoc.org/github.com/goccy/go-json?status.svg)](https://pkg.go.dev/github.com/goccy/go-json?tab=doc)
[![codecov](https://codecov.io/gh/goccy/go-json/branch/master/graph/badge.svg)](https://codecov.io/gh/goccy/go-json)

Easiest AMF encoder/decoder for Go.

# Features

- Usage exactly similar to std.json
- Flexible customization with options

# Installation

```
go get github.com/gnolizuh/gamf
```

# How to use

```
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

bs, _ := Marshal(&Struct{})

s := Struct{}
Unmarshal(bs, &s)
```

# Reference

- https://rtmp.veriskope.com/pdf/amf0-file-format-specification.pdf