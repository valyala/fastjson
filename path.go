package fastjson

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Path represents a path to a value in a JSON object.
// It is a sequence of keys (strings) and indexes (integers).
// For example, Path{"a", 0, "b"} for accessing field 'b' of the first element of array field 'a'.
type Path []interface{}

// SetP sets a value at the specified path.
// If the path does not exist, it will be created.
// special case: if path contains -1 as an index, a new array item will be added.
func (v *Value) SetP(path Path, value *Value) {
	if v == nil || len(path) == 0 {
		return
	}

	key := path[0]
	rest := path[1:]

	switch v.t {
	case TypeObject:
		k, ok := key.(string)
		if !ok {
			return
		}
		if len(rest) == 0 {
			v.o.Set(k, value)
			return
		}
		child := v.o.Get(k)
		if child == nil {
			if _, nextIsInt := rest[0].(int); nextIsInt {
				child = &Value{t: TypeArray} // create new array
			} else {
				child = &Value{t: TypeObject} // create new object
			}
			v.o.Set(k, child)
		}
		v.o.Get(k).SetP(rest, value) // recursive call

	case TypeArray:
		idx, ok := key.(int)
		if !ok {
			return
		}
		if idx == -1 {
			idx = len(v.a)
		}
		if len(rest) == 0 {
			v.SetArrayItem(idx, value)
			return
		}
		if idx >= len(v.a) { // index out of range, create new empty arr/obj
			var child *Value
			if _, nextIsInt := rest[0].(int); nextIsInt {
				child = &Value{t: TypeArray} // create new array
			} else {
				child = &Value{t: TypeObject} // create new object
			}
			v.SetArrayItem(idx, child)
		}
		v.a[idx].SetP(rest, value) // recursive call
	}
}

func (v *Value) GetP(path Path) *Value {
	if v == nil || len(path) == 0 {
		return nil
	}

	key := path[0]
	rest := path[1:]

	switch v.t {
	case TypeObject:
		k, ok := key.(string)
		if !ok {
			return nil
		}
		if len(rest) == 0 {
			return v.o.Get(k)
		}
		child := v.o.Get(k)
		if child == nil {
			return nil
		}
		return child.GetP(rest) // recursive call

	case TypeArray:
		idx, ok := key.(int)
		if !ok {
			return nil
		}
		if idx < 0 || idx >= len(v.a) {
			return nil
		}
		if len(rest) == 0 {
			return v.a[idx]
		}
		child := v.a[idx]
		if child == nil {
			return nil
		}
		return child.GetP(rest) // recursive call
	}
	return nil
}

func (v *Value) SetAny(path Path, anyVal interface{}) {
	v.SetP(path, createValueFromAny(anyVal))
}

func createValueFromAny(anyVal interface{}) *Value {
	switch v := anyVal.(type) {
	// supported scalar types defined here
	case string:
		return &Value{
			t: TypeString,
			s: v,
		}
	case int, int64, int32, int16, int8, float64, float32, uint, uint64, uint32, uint16, uint8:
		return &Value{
			t: TypeNumber,
			s: fmt.Sprintf("%v", v), // todo find a better way to convert to string
		}
	case bool:
		if v {
			return valueTrue
		} else {
			return valueFalse
		}
	case nil:
		return valueNull
	case *Value:
		return v
	case Value:
		return &v
	default:
		// use reflection to handle structs, slices, maps
		rv := reflect.ValueOf(anyVal)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		switch rv.Kind() {
		case reflect.Struct:
			obj := Object{}
			for i := 0; i < rv.NumField(); i++ {
				field := rv.Type().Field(i)
				if field.PkgPath != "" { // skip unexported field
					continue
				}
				// respect json tag if present
				tag := field.Tag.Get("json")
				if tag == "-" {
					continue
				}
				var name = field.Name
				var omitempty = false
				if tag != "" {
					name, _, _ = strings.Cut(tag, ",")
					omitempty = strings.Contains(tag, "omitempty")
				}
				if omitempty && reflect.DeepEqual(rv.Field(i).Interface(), reflect.Zero(field.Type).Interface()) {
					continue
				}
				obj.Set(name, createValueFromAny(rv.Field(i).Interface())) // recursive call
			}

			return &Value{
				t: TypeObject,
				o: obj,
			}
		case reflect.Slice:
			value := &Value{t: TypeArray}
			for i := 0; i < rv.Len(); i++ {
				value.a = append(value.a, createValueFromAny(rv.Index(i).Interface())) // recursive call
			}
			return value
		case reflect.Map:
			obj := Object{}
			for _, k := range rv.MapKeys() {
				obj.Set(k.String(), createValueFromAny(rv.MapIndex(k).Interface())) // recursive call
			}
			// sort keys alphabetically, because foreach on map is not guaranteed to be in order
			sort.Slice(obj.kvs, func(i, j int) bool {
				return obj.kvs[i].k < obj.kvs[j].k
			})

			return &Value{
				t: TypeObject,
				o: obj,
			}
		default:
			// todo implement fallback for other types
			panic(fmt.Sprintf("unsupported type: %T", anyVal))
		}
	}
}
