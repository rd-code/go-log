package log

import (
	"bytes"
	"errors"
	"testing"
)

type fieldElement struct {
	key      string
	value    any
	expected string
}

func TestIntField(t *testing.T) {
	items := []fieldElement{
		fieldElement{
			key:      "k1",
			value:    int64(123),
			expected: `"k1": 123`,
		},
		{
			key:      "k2",
			value:    int64(-123),
			expected: `"k2": -123`,
		},
		{
			key:      "k3",
			value:    int64(0),
			expected: `"k3": 0`,
		},
	}
	for _, item := range items {
		buffer := &bytes.Buffer{}
		IntField(item.key, item.value.(int64)).addToBuffer(buffer)
		if item.expected != buffer.String() {
			t.Fatalf("check [[IntField]] func failed, key:%s, value:%+v, actual:%s, expected:%s", item.key, item.value, buffer.String(), item.expected)
		}
	}
	t.Log("check [[IntField]] func succeed!")
}

func TestWriteToBuffer(t *testing.T) {
	type Element struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	buffer := &bytes.Buffer{}
	WriteToBuffer(buffer, IntField("k1", -10), UintField("k2", uint(10)), BoolField("k3", true),
		FloatField("k4", -1.34), ComplexField("k5", complex(-1, -10)), StringField("k60", "test0"),
		StringField("k61", "test1"), JsonField("k7", &Element{
			Name: "kris",
			Age:  22,
		}), ErrorField(errors.New("test error")))
	expected := `{"k1": -10, "k2": 10, "k3": true, "k4": -1.34, "k5": "(-1-10i)", "k60": "test0", "k61": "test1", "k7": {"name":"kris","age":22}, "error": "test error"}`
	if expected != buffer.String() {
		t.Fatalf("check [[WriteToBuffer]] func failed, expected:%s, actual:%s\n", expected, buffer.String())
	}
	t.Log("check [[WriteToBuffer]] func succeed!")
}
