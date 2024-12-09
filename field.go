package log

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

type fieldType int

const (
	intType = iota
	uintType
	boolType
	floatType
	complexType
	stringType
	jsonType
)

type fieldValue interface {
	AddToBuffer(buffer *bytes.Buffer)
}

type intFieldValue struct {
	value int64
}

func (i *intFieldValue) AddToBuffer(buffer *bytes.Buffer) {
	buffer.WriteString(strconv.FormatInt(i.value, 10))
}

type uintFieldValue struct {
	value uint64
}

func (i *uintFieldValue) AddToBuffer(buffer *bytes.Buffer) {
	buffer.WriteString(strconv.FormatUint(i.value, 10))
}

type boolFieldValue struct {
	value bool
}

func (b *boolFieldValue) AddToBuffer(buffer *bytes.Buffer) {
	buffer.WriteString(strconv.FormatBool(b.value))
}

type floatFieldValue struct {
	value float64
}

func (f *floatFieldValue) AddToBuffer(buffer *bytes.Buffer) {
	buffer.WriteString(strconv.FormatFloat(f.value, 'f', -1, 64))
}

type complexFieldValue struct {
	value complex128
}

func (f *complexFieldValue) AddToBuffer(buffer *bytes.Buffer) {
	buffer.WriteString(`"` + strconv.FormatComplex(f.value, 'f', -1, 128) + `"`)
}

type stringFieldValue struct {
	value string
}

func (s *stringFieldValue) AddToBuffer(buffer *bytes.Buffer) {
	buffer.WriteString(`"` + s.value + `"`)
}

type jsonFieldValue struct {
	value any
}

func (j *jsonFieldValue) AddToBuffer(buffer *bytes.Buffer) {
	byteArray, err := json.Marshal(j.value)
	if err != nil {
		fmt.Printf("encode object to json string failed, value:%+v, err:%v\n", j.value, err)
		buffer.WriteString(fmt.Sprintf(`"encode object to json failed, value:%+v, err:%v"`,
			j.value, err))
	}
	buffer.Write(byteArray)
}

type unknownFieldValue struct {
	value any
}

func (u *unknownFieldValue) AddToBuffer(buffer *bytes.Buffer) {
	buffer.WriteString(fmt.Sprintf(`"unknown type:%+v"`, u.value))
}

type fieldValueFactory interface {
	createFieldValue(typ fieldType, value any) fieldValue
}

type defaultFieldValueFactory struct {
}

func (*defaultFieldValueFactory) createFieldValue(typ fieldType, value any) fieldValue {
	switch typ {
	case intType:
		return &intFieldValue{value: value.(int64)}
	case uintType:
		return &uintFieldValue{value: value.(uint64)}
	case boolType:
		return &boolFieldValue{value: value.(bool)}
	case floatType:
		return &floatFieldValue{value: value.(float64)}
	case complexType:
		return &complexFieldValue{value: value.(complex128)}
	case stringType:
		return &stringFieldValue{value: value.(string)}
	case jsonType:
		return &jsonFieldValue{value: value}
	}
	return &unknownFieldValue{value: value}
}

func newFieldValueFactory() fieldValueFactory {
	return &defaultFieldValueFactory{}
}

type Field struct {
	key   string
	value fieldValue
}

func (f *Field) addToBuffer(buffer *bytes.Buffer) {
	buffer.WriteByte('"')
	buffer.WriteString(f.key)
	buffer.WriteString(`": `)
	f.value.AddToBuffer(buffer)
}

func NewField(key string, typ fieldType, value any) *Field {
	return &Field{
		key:   key,
		value: newFieldValueFactory().createFieldValue(typ, value),
	}
}

func UintField[T uint8 | uint16 | uint32 | uint64 | uint](key string, value T) *Field {
	return NewField(key, uintType, uint64(value))
}

func IntField[T int8 | int16 | int32 | int64 | int](key string, value T) *Field {
	return NewField(key, intType, int64(value))
}

func BoolField(key string, value bool) *Field {
	return NewField(key, boolType, value)
}

func FloatField[T float32 | float64](key string, value T) *Field {
	return NewField(key, floatType, float64(value))
}

func ComplexField[T complex64 | complex128](key string, value T) *Field {
	return NewField(key, complexType, complex128(value))
}

func StringField(key, value string) *Field {
	return NewField(key, stringType, value)
}

func JsonField(key string, value any) *Field {
	return NewField(key, jsonType, value)
}

func ErrorField(err error) *Field {
	return NewField("error", stringType, err.Error())
}

// 按照json格式写入到buffer中
func WriteToBuffer(buffer *bytes.Buffer, fields ...*Field) {
	buffer.WriteByte('{')
	if len(fields) == 0 {
		goto label
	}
	fields[0].addToBuffer(buffer)
	for i := 1; i < len(fields); i++ {
		buffer.WriteString(", ")
		fields[i].addToBuffer(buffer)
	}
label:
	buffer.WriteByte('}')
}
