package fastjson

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
