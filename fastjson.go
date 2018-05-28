package fastjson

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// Parser parses JSON.
//
// Parser may be re-used for subsequent parsing.
//
// Parser cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Parser struct {
	// b contains the parsed string.
	b []byte

	// v contains the parsed value.
	v *Value

	// c is a cache for json values.
	c cache
}

// Parse parses s containing JSON.
//
// The parsed value may be obtained via Value call.
func (p *Parser) Parse(s string) error {
	s = skipWS(s)
	p.b = append(p.b[:0], s...)
	p.c.reset()

	v, tail, err := parseValue(b2s(p.b), &p.c)
	if err != nil {
		return fmt.Errorf("cannot parse JSON: %s; unparsed tail: %q", err, tail)
	}
	tail = skipWS(tail)
	if len(tail) > 0 {
		return fmt.Errorf("unexpected tail: %q", tail)
	}
	p.v = v
	return nil
}

// ParseBytes parses b containing JSON.
//
// The parsed value may be obtained via Value call.
func (p *Parser) ParseBytes(b []byte) error {
	return p.Parse(b2s(b))
}

// Value returns the parsed value.
//
// The parsed value is valid until the next call to Parse.
func (p *Parser) Value() *Value {
	return p.v
}

type cache struct {
	vs []Value
}

func (c *cache) reset() {
	c.vs = c.vs[:0]
}

func (c *cache) getValue() *Value {
	if cap(c.vs) > len(c.vs) {
		c.vs = c.vs[:len(c.vs)+1]
	} else {
		c.vs = append(c.vs, Value{})
	}
	v := &c.vs[len(c.vs)-1]
	v.reset()
	return v
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	strh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	var sh reflect.SliceHeader
	sh.Data = strh.Data
	sh.Len = strh.Len
	sh.Cap = strh.Len
	return *(*[]byte)(unsafe.Pointer(&sh))
}

func skipWS(s string) string {
	if len(s) == 0 || s[0] > '\x20' {
		// Fast path.
		return s
	}

	// Slow path.
	for i := 0; i < len(s); i++ {
		switch s[i] {
		// Whitespace chars are obtained from http://www.ietf.org/rfc/rfc4627.txt .
		case '\x20', '\x0D', '\x0A', '\x09':
			continue
		default:
			return s[i:]
		}
	}
	return ""
}

type kv struct {
	k string
	v *Value
}

func parseValue(s string, c *cache) (*Value, string, error) {
	if len(s) == 0 {
		return nil, s, fmt.Errorf("cannot parse empty string")
	}

	var v *Value
	var err error

	switch s[0] {
	case '{':
		v, s, err = parseObject(s, c)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse object: %s", err)
		}
		return v, s, nil
	case '[':
		v, s, err = parseArray(s, c)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse array: %s", err)
		}
		return v, s, nil
	case '"':
		var ss string
		ss, s, err = parseRawString(s)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse string: %s", err)
		}
		v = c.getValue()
		v.t = typeRawString
		v.s = ss
		return v, s, nil
	case 't':
		if !strings.HasPrefix(s, "true") {
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		s = s[len("true"):]
		return valueTrue, s, nil
	case 'f':
		if !strings.HasPrefix(s, "false") {
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		s = s[len("false"):]
		return valueFalse, s, nil
	case 'n':
		if !strings.HasPrefix(s, "null") {
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		s = s[len("null"):]
		return valueNull, s, nil
	default:
		var ns string
		ns, s, err = parseRawNumber(s)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse number: %s", err)
		}
		v = c.getValue()
		v.t = typeRawNumber
		v.s = ns
		return v, s, nil
	}
}

func parseArray(s string, c *cache) (*Value, string, error) {
	// Skip the first char - '['
	s = s[1:]

	s = skipWS(s)
	if len(s) == 0 {
		return nil, s, fmt.Errorf("missing ']'")
	}

	if s[0] == ']' {
		return emptyArray, s[1:], nil
	}

	a := c.getValue()
	a.t = TypeArray
	for {
		var v *Value
		var err error

		s = skipWS(s)
		v, s, err = parseValue(s, c)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse array value: %s", err)
		}
		a.a = append(a.a, v)

		s = skipWS(s)
		if len(s) == 0 {
			return nil, s, fmt.Errorf("unexpected end of array")
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		if s[0] == ']' {
			s = s[1:]
			return a, s, nil
		}
		return nil, s, fmt.Errorf("missing ',' after array value")
	}
}

func parseObject(s string, c *cache) (*Value, string, error) {
	// Skip the first char - '{'
	s = s[1:]

	s = skipWS(s)
	if len(s) == 0 {
		return nil, s, fmt.Errorf("missing '}'")
	}

	if s[0] == '}' {
		return emptyObject, s[1:], nil
	}

	o := c.getValue()
	o.t = TypeObject
	for {
		var err error
		kv := o.o.getKV()

		// Parse key.
		s = skipWS(s)
		kv.k, s, err = parseRawString(s)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse object key: %s", err)
		}
		s = skipWS(s)
		if len(s) == 0 || s[0] != ':' {
			return nil, s, fmt.Errorf("missing ':' after object key")
		}
		s = s[1:]

		// Parse value
		s = skipWS(s)
		kv.v, s, err = parseValue(s, c)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parse object value: %s", err)
		}
		s = skipWS(s)
		if len(s) == 0 {
			return nil, s, fmt.Errorf("unexpected end of object")
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		if s[0] == '}' {
			return o, s[1:], nil
		}
		return nil, s, fmt.Errorf("missing ',' after object value")
	}
}

func unescapeStringBestEffort(s string) string {
	n := strings.IndexByte(s, '\\')
	if n < 0 {
		// Fast path - nothing to unescape.
		return s
	}

	// Slow path - unescape string.
	b := s2b(s) // It is safe to do, since s points to a byte slice in Parser.b.
	b = b[:n]
	s = s[n+1:]
	for len(s) > 0 {
		ch := s[0]
		s = s[1:]
		switch ch {
		case '"':
			b = append(b, '"')
		case '\\':
			b = append(b, '\\')
		case '/':
			b = append(b, '/')
		case 'b':
			b = append(b, '\b')
		case 'f':
			b = append(b, '\f')
		case 'n':
			b = append(b, '\n')
		case 'r':
			b = append(b, '\r')
		case 't':
			b = append(b, '\t')
		case 'u':
			if len(s) < 4 {
				return b2s(b)
			}
			xs := s[:4]
			x, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				return b2s(b)
			}
			b = append(b, string(rune(x))...)
			s = s[4:]
		}
		n = strings.IndexByte(s, '\\')
		if n < 0 {
			b = append(b, s...)
			break
		}
		b = append(b, s[:n]...)
		s = s[n+1:]
	}
	return b2s(b)
}

func parseRawString(s string) (string, string, error) {
	if len(s) == 0 || s[0] != '"' {
		return "", s, fmt.Errorf(`missing opening '"'`)
	}
	s = s[1:]
	ss := s

	for {
		n := strings.IndexByte(s, '"')
		if n < 0 {
			return "", s, fmt.Errorf(`missing closing '"'`)
		}
		if n == 0 || s[n-1] != '\\' {
			return ss[:len(ss)-len(s)+n], s[n+1:], nil
		}

		i := n - 2
		for i > 0 && s[i] == '\\' {
			i--
		}
		if uint(n-i)%2 == 1 {
			return ss[:len(ss)-len(s)+n], s[n+1:], nil
		}
		s = s[n+1:]
	}
}

func parseRawNumber(s string) (string, string, error) {
	ch := s[0]
	if ch != '-' && (ch < '0' || ch > '9') {
		return "", s, fmt.Errorf("unexpected chars in the number: %q", s)
	}

	// Find the end of the number.
	n := len(s)
	for i := 1; i < len(s); i++ {
		ch := s[i]
		if ch == '-' || (ch >= '0' && ch <= '9') || ch == '.' || ch == 'e' || ch == 'E' || ch == '+' {
			continue
		}
		n = i
		break
	}

	ns := s[:n]
	s = s[n:]
	return ns, s, nil
}

func parseNumberOrZero(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

// Object represents JSON object.
//
// Object cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Object struct {
	kvs           []kv
	keysUnescaped bool
}

func (o *Object) reset() {
	o.kvs = o.kvs[:0]
	o.keysUnescaped = false
}

// String returns string representation for the o.
//
// This function is for debugging purposes only. It isn't optimized for speed.
func (o *Object) String() string {
	o.unescapeKeys()

	var sb strings.Builder
	fmt.Fprintf(&sb, "{")
	for i, kv := range o.kvs {
		fmt.Fprintf(&sb, "%q:%s", kv.k, kv.v)
		if i != len(o.kvs)-1 {
			fmt.Fprintf(&sb, ",")
		}
	}
	fmt.Fprintf(&sb, "}")
	return sb.String()
}

func (o *Object) getKV() *kv {
	if cap(o.kvs) > len(o.kvs) {
		o.kvs = o.kvs[:len(o.kvs)+1]
	} else {
		o.kvs = append(o.kvs, kv{})
	}
	return &o.kvs[len(o.kvs)-1]
}

func (o *Object) unescapeKeys() {
	if o.keysUnescaped {
		return
	}
	for i := range o.kvs {
		kv := &o.kvs[i]
		kv.k = unescapeStringBestEffort(kv.k)
	}
	o.keysUnescaped = true
}

// Len returns the number of items in the o.
func (o *Object) Len() int {
	return len(o.kvs)
}

// Get returns the value for the given key in the o.
//
// Returns nil if the value for the given key isn't found.
//
// The returned value is valid until Parse is called on the Parser returned o.
func (o *Object) Get(key string) *Value {
	o.unescapeKeys()

	for i := range o.kvs {
		kv := &o.kvs[i]
		if string(kv.k) == key {
			return kv.v
		}
	}
	return nil
}

// Visit calls f for each item in the o.
//
// f cannot hold key and/or v after returning.
func (o *Object) Visit(f func(key string, v *Value)) {
	o.unescapeKeys()

	for i := range o.kvs {
		kv := &o.kvs[i]
		f(kv.k, kv.v)
	}
}

// Value represents any JSON value.
//
// Use Type in order to determine the actual type of the JSON value.
// Use Get* for obtaining actual values.
//
// Value cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Value struct {
	o Object
	a []*Value
	s string
	n float64
	t Type
}

func (v *Value) reset() {
	v.o.reset()
	v.a = v.a[:0]
	v.s = ""
	v.n = 0
	v.t = TypeNull
}

// String returns string representation of the v.
//
// The function is for debugging purposes only. It isn't optimized for speed.
func (v *Value) String() string {
	switch v.t {
	case typeRawString:
		v.s = unescapeStringBestEffort(v.s)
		v.t = TypeString
	case typeRawNumber:
		v.n = parseNumberOrZero(v.s)
		v.t = TypeNumber
	}

	switch v.t {
	case TypeObject:
		return v.o.String()
	case TypeArray:
		var sb strings.Builder
		fmt.Fprintf(&sb, "[")
		for i, vv := range v.a {
			fmt.Fprintf(&sb, "%s", vv)
			if i != len(v.a)-1 {
				fmt.Fprintf(&sb, ",")
			}
		}
		fmt.Fprintf(&sb, "]")
		return sb.String()
	case TypeString:
		return fmt.Sprintf("%q", v.s)
	case TypeNumber:
		if float64(int(v.n)) == v.n {
			return fmt.Sprintf("%d", int(v.n))
		}
		return fmt.Sprintf("%f", v.n)
	case TypeTrue:
		return "true"
	case TypeFalse:
		return "false"
	case TypeNull:
		return "null"
	default:
		panic(fmt.Errorf("BUG: unknown Value type: %d", v.t))
	}
}

// Type represents JSON type.
type Type int

const (
	// TypeObject is JSON object type.
	TypeObject = Type(0)

	// TypeArray is JSON array type.
	TypeArray = Type(1)

	// TypeString is JSON string type.
	TypeString = Type(2)

	// TypeNumber is JSON number type.
	TypeNumber = Type(3)

	// TypeTrue is JSON true.
	TypeTrue = Type(4)

	// TypeFalse is JSON false.
	TypeFalse = Type(5)

	// TypeNull is JSON null.
	TypeNull = Type(6)

	typeRawString = Type(7)
	typeRawNumber = Type(8)
)

// String returns string representation of t.
func (t Type) String() string {
	switch t {
	case TypeObject:
		return "object"
	case TypeArray:
		return "array"
	case TypeString:
		return "string"
	case TypeNumber:
		return "number"
	case TypeTrue:
		return "true"
	case TypeFalse:
		return "false"
	case TypeNull:
		return "null"

	// typeRawString and typeRawNumber are skipped intentionally,
	// since they shouldn't be visible to user.
	default:
		panic(fmt.Errorf("BUG: unknown Value type: %d", t))
	}
}

// Type returns the type of the v.
func (v *Value) Type() Type {
	switch v.t {
	case typeRawString:
		v.s = unescapeStringBestEffort(v.s)
		v.t = TypeString
	case typeRawNumber:
		v.n = parseNumberOrZero(v.s)
		v.t = TypeNumber
	}
	return v.t
}

// Get returns value by the given keys path.
//
// Array indexes may be represented as decimal numbers.
//
// nil is returned for non-existing keys path.
//
// The returned value is valid until Parse is called on the Parser returned v.
func (v *Value) Get(keys ...string) *Value {
	for _, key := range keys {
		switch v.t {
		case TypeObject:
			v = v.o.Get(key)
			if v == nil {
				return nil
			}
		case TypeArray:
			n, err := strconv.Atoi(key)
			if err != nil || n < 0 || n >= len(v.a) {
				return nil
			}
			v = v.a[n]
		default:
			return nil
		}
	}
	return v
}

// GetFloat64 returns float64 value by the given keys path.
//
// Array indexes may be represented as decimal numbers.
//
// 0 is returned for non-existing keys path or invalid value type.
func (v *Value) GetFloat64(keys ...string) float64 {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	return v.Float64()
}

// GetInt returns int value by the given keys path.
//
// Array indexes may be represented as decimal numbers.
//
// 0 is returned for non-existing keys path or invalid value type.
func (v *Value) GetInt(keys ...string) int {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	return v.Int()
}

// GetStringBytes returns string value by the given keys path.
//
// Array indexes may be represented as decimal numbers.
//
// nil is returned for non-existing keys path or invalid value type.
func (v *Value) GetStringBytes(keys ...string) []byte {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeString {
		return nil
	}
	return v.StringBytes()
}

// GetBool returns bool value by the given keys path.
//
// Array indexes may be represented as decimal numbers.
//
// false is returned for non-existing keys path or invalid value type.
func (v *Value) GetBool(keys ...string) bool {
	v = v.Get(keys...)
	if v != nil && v.Type() == TypeTrue {
		return true
	}
	return false
}

// Object returns the underlying JSON object for the v.
//
// The returned object is valid until Parse is called on the Parser returned v.
func (v *Value) Object() *Object {
	if v.t != TypeObject {
		panic(fmt.Errorf("BUG: value doesn't contain object; it contains %s", v.t))
	}
	return &v.o
}

// Array returns the underlying JSON array for the v.
//
// The returned array is valid until Parse is called on the Parser returned v.
func (v *Value) Array() []*Value {
	if v.t != TypeArray {
		panic(fmt.Errorf("BUG: value doesn't contain array; it contains %s", v.t))
	}
	return v.a
}

// StringBytes returns the underlying JSON string for the v.
//
// The returned string is valid until Parse is called on the Parser returned v.
func (v *Value) StringBytes() []byte {
	if v.t == typeRawString {
		v.s = unescapeStringBestEffort(v.s)
		v.t = TypeString
	}
	if v.t != TypeString {
		panic(fmt.Errorf("BUG: value doesn't contain string; it contains %s", v.t))
	}
	return s2b(v.s)
}

// Float64 returns the underlying JSON number for the v.
func (v *Value) Float64() float64 {
	if v.t == typeRawNumber {
		v.n = parseNumberOrZero(v.s)
		v.t = TypeNumber
	}
	if v.t != TypeNumber {
		panic(fmt.Errorf("BUG: value doesn't contain number; it contains %s", v.t))
	}
	return v.n
}

// Int returns the underlying JSON int for the v.
func (v *Value) Int() int {
	f := v.Float64()
	return int(f)
}

var (
	valueTrue   = &Value{t: TypeTrue}
	valueFalse  = &Value{t: TypeFalse}
	valueNull   = &Value{t: TypeNull}
	emptyObject = &Value{t: TypeObject}
	emptyArray  = &Value{t: TypeArray}
)
