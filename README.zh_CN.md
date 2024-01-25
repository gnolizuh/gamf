[English](README.md) | 中文

# gamf

![Go](https://github.com/goccy/go-json/workflows/Go/badge.svg)
[![GoDoc](https://godoc.org/github.com/goccy/go-json?status.svg)](https://pkg.go.dev/github.com/goccy/go-json?tab=doc)
[![codecov](https://codecov.io/gh/goccy/go-json/branch/master/graph/badge.svg)](https://codecov.io/gh/goccy/go-json)

基于Go实现的极简AMF编解码库

# 特性

- 用法与json标准库类似
- 灵活的自定义选项

# 安装

```
go get github.com/gnolizuh/gamf
```

# 用法

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

# 引用

- https://rtmp.veriskope.com/pdf/amf0-file-format-specification.pdf