package fastjson

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCreateObjectParser(t *testing.T) {
	var p Parser
	var pi float64 = 3.1415926535898
	v := p.NewObject(map[string]*Value{
		"foo":       p.NewString("bar"),
		"bool_true": p.NewBool(true),
		"number":    p.NewFloat64(pi),
		"array": p.NewArray([]*Value{
			p.NewString("hello"),
			p.NewInt(42),
			p.NewBool(false),
		}),
		"null": p.NewNull(),
	})
	s := v.MarshalTo(nil)
	var val interface{}
	err := json.Unmarshal(s, &val)
	t.Log(string(s))
	if err != nil {
		t.Fatalf("Cannot unmarshal json: %s", err)
	}
	var expectedVal = map[string]interface{}{
		"foo":       "bar",
		"bool_true": true,
		"number":    pi,
		"array": []interface{}{
			"hello",
			float64(42),
			false,
		},
		"null": nil,
	}
	if !reflect.DeepEqual(val, expectedVal) {
		t.Fatal("JSON result does not match expected result")
	}
}

func TestCreateObject(t *testing.T) {
	var pi float64 = 3.1415926535898
	v := NewObject(map[string]*Value{
		"foo":       NewString("bar"),
		"bool_true": NewBool(true),
		"number":    NewFloat64(pi),
		"array": NewArray([]*Value{
			NewString("hello"),
			NewInt(42),
			NewBool(false),
		}),
		"null": NewNull(),
	})
	s := v.MarshalTo(nil)
	var val interface{}
	err := json.Unmarshal(s, &val)
	t.Log(string(s))
	if err != nil {
		t.Fatalf("Cannot unmarshal json: %s", err)
	}
	var expectedVal = map[string]interface{}{
		"foo":       "bar",
		"bool_true": true,
		"number":    pi,
		"array": []interface{}{
			"hello",
			float64(42),
			false,
		},
		"null": nil,
	}
	if !reflect.DeepEqual(val, expectedVal) {
		t.Fatal("JSON result does not match expected result")
	}
}
