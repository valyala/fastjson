package fastjson

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCreateObject(t *testing.T) {
	var p Parser
	var pi float64 = 3.1415926535898
	v := p.NewObject(map[string]*Value{
		"foo":       p.NewString("bar"),
		"bool_true": p.NewBool(true),
		"number":    p.NewFloat64(pi),
		"array": p.NewArray([]*Value{
			p.NewString("hello"),
			p.NewBool(false),
		}),
	})
	s := v.MarshalTo(nil)
	var val interface{}
	err := json.Unmarshal(s, &val)
	if err != nil {
		t.Fatalf("Cannot unmarshal json: %s", err)
	}
	var expectedVal = map[string]interface{}{
		"foo":       "bar",
		"bool_true": true,
		"number":    pi,
		"array": []interface{}{
			"hello",
			false,
		},
	}
	if !reflect.DeepEqual(val, expectedVal) {
		t.Fatal("JSON result does not match expected result")
	}
}
