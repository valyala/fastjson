package fastjson

import (
	"testing"
)

func TestValue_SetP(t *testing.T) {
	t.Run("Update existing nested value", func(t *testing.T) {
		v := MustParse(`{"a": {"b": 1}}`)
		v.SetP(Path{"a", "b"}, MustParse("2"))
		if val := v.Get("a").Get("b").GetInt(); val != 2 {
			t.Fatalf("expected 2, got %v", val)
		}
	})

	t.Run("Set new top-level key", func(t *testing.T) {
		v := MustParse(`{}`)
		v.Set("z", MustParse(`111`))
		if val := v.Get("z").GetInt(); val != 111 {
			t.Fatalf("expected 111, got %v", val)
		}
	})

	t.Run("Set new nested key in object", func(t *testing.T) {
		v := MustParse(`{"a": {}}`)
		v.SetP(Path{"a", "bb"}, MustParse(`555`))
		if val := v.Get("a").Get("bb").GetInt(); val != 555 {
			t.Fatalf("expected 555, got %v", val)
		}
	})

	t.Run("Set deep nested key with intermediate objects", func(t *testing.T) {
		v := MustParse(`{"a": {}}`)
		v.SetP(Path{"a", "sub", "subsub", "s3"}, MustParse(`666`))
		if val := v.Get("a").Get("sub").Get("subsub").Get("s3").GetInt(); val != 666 {
			t.Fatalf("expected 666, got %v", val)
		}
	})

	t.Run("Update array element by index", func(t *testing.T) {
		v := MustParse(`{"arr": [99]}`)
		v.SetP(Path{"arr", 0}, MustParse(`111`))
		if val := v.Get("arr").Get("0").GetInt(); val != 111 {
			t.Fatalf("expected 111, got %v", val)
		}
	})

	t.Run("Add element to empty array", func(t *testing.T) {
		v := MustParse(`{"arr": []}`)
		v.SetP(Path{"arr", 0}, MustParse(`111`))
		if val := v.Get("arr").Get("0").GetInt(); val != 111 {
			t.Fatalf("expected 111, got %v", val)
		}
	})

	// Test attempting to set nested value in non-array (operation should have no effect).
	t.Run("Attempt to set nested value in non-array", func(t *testing.T) {
		v := MustParse(`{"a":[0]}`)
		v.SetP(Path{"a", 0, 0}, MustParse(`222`))
		if v.String() != `{"a":[0]}` {
			t.Fatalf("expected unchanged value, got %s", v.String())
		}
	})

	// Test creating missing null values in array when setting element beyond current length.
	t.Run("Create missing null values in array", func(t *testing.T) {
		v := MustParse(`{"a":[0]}`)
		v.SetP(Path{"a", 2}, MustParse(`111`))
		if val := v.Get("a").Get("0").GetInt(); val != 0 {
			t.Fatalf("expected 0 at index 0, got %v", val)
		}
		if typ := v.Get("a").Get("1").Type(); typ != TypeNull {
			t.Fatalf("expected null at index 1, got type %v", typ)
		}
		if val := v.Get("a").Get("2").GetInt(); val != 111 {
			t.Fatalf("expected 111 at index 2, got %v", val)
		}
	})

	// Test -1 index to append to array.
	t.Run("Append to array with -1 index", func(t *testing.T) {
		v := MustParse(`[]`)
		v.SetP(Path{-1}, MustParse(`111`))
		if v.String() != `[111]` {
			t.Fatalf("expected appended value, got %s", v.String())
		}

		v = MustParse(`{"a":[0]}`)
		v.SetP(Path{"a", -1}, MustParse(`111`))
		if v.String() != `{"a":[0,111]}` {
			t.Fatalf("expected appended value, got %s", v.String())
		}
	})

	// Test -1 index in the middle of the path.
	t.Run("Append to array with -1 index in the middle of the path", func(t *testing.T) {
		v := MustParse(`{}`)
		v.SetP(Path{"a", -1, "aa"}, MustParse(`1`))
		v.SetP(Path{"a", -1, "bb"}, MustParse(`2`))
		v.SetP(Path{"a", -1, "cc"}, MustParse(`3`))
		if v.String() != `{"a":[{"aa":1},{"bb":2},{"cc":3}]}` {
			t.Fatalf("expected appended value, got %s", v.String())
		}
	})
}

func TestValue_GetP(t *testing.T) {
	t.Run("Get existing nested value", func(t *testing.T) {
		v := MustParse(`{"a": {"b": 1}}`)
		if val := v.GetP(Path{"a", "b"}).GetInt(); val != 1 {
			t.Fatalf("expected 1, got %v", val)
		}
	})

	t.Run("Get non-existing nested value", func(t *testing.T) {
		v := MustParse(`{"a": {"b": 1}}`)
		if val := v.GetP(Path{"a", "c"}); val != nil {
			t.Fatalf("expected nil, got %v", val)
		}
	})

	t.Run("Get value from array by index", func(t *testing.T) {
		v := MustParse(`[1,2,3,[4,5,6,[7,8,9]]]`)
		if val := v.GetP(Path{3, 2}).GetInt(); val != 6 {
			t.Fatalf("expected 6, got %v", val)
		}
		if val := v.GetP(Path{3, 3, 1}).GetInt(); val != 8 {
			t.Fatalf("expected 8, got %v", val)
		}
	})

	t.Run("Get from array inside object", func(t *testing.T) {
		v := MustParse(`{"a": [1,2,3]}`)
		if val := v.GetP(Path{"a", 1}).GetInt(); val != 2 {
			t.Fatalf("expected 2, got %v", val)
		}
	})

	t.Run("Get from object inside array", func(t *testing.T) {
		v := MustParse(`[{"a":1},{"b":2}]`)
		if val := v.GetP(Path{1, "b"}).GetInt(); val != 2 {
			t.Fatalf("expected 2, got %v", val)
		}
	})
}
