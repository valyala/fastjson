package fastjson

import (
	"testing"
)

func TestObjectDelSet(t *testing.T) {
	var p Parser
	var o *Object

	o.Del("xx")

	v, err := p.Parse(`{"fo\no": "bar", "x": [1,2,3]}`)
	if err != nil {
		t.Fatalf("unexpected error during parse: %s", err)
	}
	o, err = v.Object()
	if err != nil {
		t.Fatalf("cannot obtain object: %s", err)
	}

	// Delete x
	o.Del("x")
	if o.Len() != 1 {
		t.Fatalf("unexpected number of items left; got %d; want %d", o.Len(), 1)
	}

	// Try deleting non-existing value
	o.Del("xxx")
	if o.Len() != 1 {
		t.Fatalf("unexpected number of items left; got %d; want %d", o.Len(), 1)
	}

	// Set new value
	vNew := MustParse(`{"foo":[1,2,3]}`)
	o.Set("new_key", vNew)

	// Delete item with escaped key
	o.Del("fo\no")
	if o.Len() != 1 {
		t.Fatalf("unexpected number of items left; got %d; want %d", o.Len(), 1)
	}

	str := o.String()
	strExpected := `{"new_key":{"foo":[1,2,3]}}`
	if str != strExpected {
		t.Fatalf("unexpected string representation for o: got %q; want %q", str, strExpected)
	}

	// Set and Del function as no-op on nil value
	o = nil
	o.Del("x")
	o.Set("x", MustParse(`[3]`))
}

func TestValueDelSet(t *testing.T) {
	var p Parser
	v, err := p.Parse(`{"xx": 123, "x": [1,2,3]}`)
	if err != nil {
		t.Fatalf("unexpected error during parse: %s", err)
	}

	// Delete xx
	v.Del("xx")
	n := v.GetObject().Len()
	if n != 1 {
		t.Fatalf("unexpected number of items left; got %d; want %d", n, 1)
	}

	// Try deleting non-existing value in the array
	va := v.Get("x")
	va.Del("foobar")

	// Delete middle element in the array
	va.Del("1")
	a := v.GetArray("x")
	if len(a) != 2 {
		t.Fatalf("unexpected number of items left in the array; got %d; want %d", len(a), 2)
	}

	// Update the first element in the array
	vNew := MustParse(`"foobar"`)
	va.Set("0", vNew)

	// Add third element to the array
	vNew = MustParse(`[3]`)
	va.Set("3", vNew)

	// Add invalid array index to the array
	va.Set("invalid", MustParse(`"nonsense"`))

	str := v.String()
	strExpected := `{"x":["foobar",3,null,[3]]}`
	if str != strExpected {
		t.Fatalf("unexpected string representation for o: got %q; want %q", str, strExpected)
	}

	// Set and Del function as no-op on nil value
	v = nil
	v.Del("x")
	v.Set("x", MustParse(`[]`))
	v.SetArrayItem(1, MustParse(`[]`))
}

func TestValueSetArrayItem(t *testing.T) {
	minInt := -int(^uint(0)>>1) - 1
	for _, tc := range []struct {
		name        string
		input       string
		idx         int
		replacement string
		want        string
	}{
		{"empty_negative", `[]`, -1, `3`, `[]`},
		{"empty_negative_nil", `[]`, -1, "", `[]`},
		{"empty_negative_two", `[]`, -2, `3`, `[]`},
		{"empty_negative_two_nil", `[]`, -2, "", `[]`},
		{"empty_min_int", `[]`, minInt, `3`, `[]`},
		{"empty_min_int_nil", `[]`, minInt, "", `[]`},
		{"populated_negative", `[1,2]`, -1, `3`, `[1,2]`},
		{"populated_negative_nil", `[1,2]`, -1, "", `[1,2]`},
		{"populated_negative_two", `[1,2]`, -2, `3`, `[1,2]`},
		{"populated_negative_two_nil", `[1,2]`, -2, "", `[1,2]`},
		{"populated_min_int", `[1,2]`, minInt, `3`, `[1,2]`},
		{"populated_min_int_nil", `[1,2]`, minInt, "", `[1,2]`},
		{"replace", `[1,2]`, 0, `3`, `[3,2]`},
		{"replace_nil", `[1,2]`, 0, "", `[null,2]`},
		{"grow", `[1,2]`, 3, `3`, `[1,2,null,3]`},
		{"grow_nil", `[1,2]`, 3, "", `[1,2,null,null]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := MustParse(tc.input)
			var replacement *Value
			if tc.replacement != "" {
				replacement = MustParse(tc.replacement)
			}
			v.SetArrayItem(tc.idx, replacement)
			if got := v.String(); got != tc.want {
				t.Fatalf("unexpected array; got %s; want %s", got, tc.want)
			}
		})
	}
}
