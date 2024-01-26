[English](README.md) | 中文

# gamf

基于Go实现的极简AMF编解码库

# 特性

- 用法与json标准库类似
- 灵活的自定义选项

# 安装

```
go get github.com/gnolizuh/gamf
```

# 用法

## Integer

```
in := 1
bs, _ := Marshal(&in)

var out int
Unmarshal(bs, &out)
```

## Float

```
in := 1.0
bs, _ := Marshal(&in)

var out float64
Unmarshal(bs, &out)
```

## String

```
in := "1"
bs, _ := Marshal(&in)

var out string
Unmarshal(bs, &out)
```

## Bool

```
in := "1"
bs, _ := Marshal(&in)

var out string
Unmarshal(bs, &out)
```

## Slice

```
in := []int{1, 2, 3}
bs, _ := Marshal(&in)

var out []int
Unmarshal(bs, &out)
```

## Struct

### Struct to Struct

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

in := Struct{} // with value be initialized
bs, _ := Marshal(&in)

out := Struct{}
Unmarshal(bs, &out)
```

### Struct to Map

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

in := Struct{} // with value be initialized
bs, _ := Marshal(&in)

out := make(map[string]interface{})
Unmarshal(bs, &out)
```

## Map

```
in := map[string]interface{}{"Int": 1.0, "String": "1", "Bool": true}
bs, _ := Marshal(&in)

out := make(map[string]interface{})
Unmarshal(bs, &out)
```

# 引用

- https://rtmp.veriskope.com/pdf/amf0-file-format-specification.pdf