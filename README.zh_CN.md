[English](README.md) | 中文

# gamf

最好用的、基于Go实现的、AMF序列化/反序列化开源库。

# 特性

- 用法与json标准库类似
- 灵活的自定义选项
- 更好的兼容性，同时支持AMF0和AMF3

# 安装

```
go get github.com/gnolizuh/gamf
```

# 用法

## Integer

### AMF0

```
in := 1
bs, _ := Marshal(&in)

var out int
Unmarshal(bs, &out)
```

### AMF3

```
var bs []byte
buf := bytes.NewBuffer(bs)

in := 1
NewEncoder(buf).Encode(&in)

var out int
NewDecoder(buf).Decode(&out)
```

## Float

### AMF0

```
in := 1.0
bs, _ := Marshal(&in)

var out float64
Unmarshal(bs, &out)
```

### AMF3

```
var bs []byte
buf := bytes.NewBuffer(bs)

in := 1
NewEncoder(buf).WithVersion3().Encode(&in)

var out int
NewDecoder(buf).WithVersion3().Decode(&out)
```

## String

### AMF0

```
in := "1"
bs, _ := Marshal(&in)

var out string
Unmarshal(bs, &out)
```

### AMF3

```
var bs []byte
buf := bytes.NewBuffer(bs)

in := "1"
NewEncoder(buf).WithVersion3().Encode(&in)

var out string
NewDecoder(buf).WithVersion3().Decode(&out)
```

## Bool

### AMF0

```
in := "1"
bs, _ := Marshal(&in)

var out string
Unmarshal(bs, &out)
```

### AMF3

```
var bs []byte
buf := bytes.NewBuffer(bs)

in := true
NewEncoder(buf).WithVersion3().Encode(&in)

out := false
NewDecoder(buf).WithVersion3().Decode(&out)
```

## Slice

### AMF0

```
in := []int{1, 2, 3}
bs, _ := Marshal(&in)

var out []int
Unmarshal(bs, &out)
```

### AMF3

```
var bs []byte
buf := bytes.NewBuffer(bs)

in := []interface{}{1.0, "1", true, map[string]interface{}{"Int": 1.0, "String": "1", "Bool": true}}
NewEncoder(buf).WithVersion3().Encode(&in)

out := []interface{}{0.0, "0", false, map[string]interface{}{}}
NewDecoder(buf).WithVersion3().Decode(&out)
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
- https://rtmp.veriskope.com/pdf/amf3-file-format-spec.pdf