package fastjson

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/valyala/fastjson/fastfloat"
)

// Parser parses JSON.
//
// Parser may be re-used for subsequent parsing.
//
// Parser cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Parser struct {
	// b contains working copy of the string to be parsed.
	b []byte

	// c is a cache for json values.
	c cache
}

// Parse parses s containing JSON.
//
// The returned value is valid until the next call to Parse*.
//
// Use Scanner if a stream of JSON values must be parsed.
func (p *Parser) Parse(s string) (*Value, error) {
	s = s[skipWS(s):]
	p.b = append(p.b[:0], s...)
	p.c.reset()

	v, tail, err := parseValue(b2s(p.b), 0, &p.c, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot parse JSON: %s; unparsed tail: %q", err, startEndString(tail))
	}
	tail = tail[skipWS(tail):]
	if len(tail) > 0 {
		return nil, fmt.Errorf("unexpected tail: %q", startEndString(tail))
	}
	return v, nil
}

// ParseBytes parses b containing JSON.
//
// The returned Value is valid until the next call to Parse*.
//
// Use Scanner if a stream of JSON values must be parsed.
func (p *Parser) ParseBytes(b []byte) (*Value, error) {
	return p.Parse(b2s(b))
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
	// Do not reset the value, since the caller must properly init it.
	return &c.vs[len(c.vs)-1]
}

func skipWS(s string) int {
	if len(s) == 0 || s[0] > 0x20 {
		// Fast path.
		return 0
	}
	return skipWSSlow(s)
}

func skipWSSlow(s string) int {
	if len(s) == 0 || s[0] != 0x20 && s[0] != 0x0A && s[0] != 0x09 && s[0] != 0x0D {
		return 0
	}
	for i := 1; i < len(s); i++ {
		if s[i] != 0x20 && s[i] != 0x0A && s[i] != 0x09 && s[i] != 0x0D {
			return i
		}
	}
	return len(s)
}

type kv struct {
	k string
	v *Value
}

// MaxDepth is the maximum depth for nested JSON.
const MaxDepth = 300

func parseValue(s string, offset int, c *cache, depth int) (*Value, string, error) {
	if offset >= len(s) {
		return nil, s[offset:], fmt.Errorf("cannot parse empty string")
	}
	depth++
	if depth > MaxDepth {
		return nil, s, fmt.Errorf("too big depth for the nested JSON; it exceeds %d", MaxDepth)
	}

	if s[offset] == '{' {
		v, olen, err := parseObject(s, offset, c, depth)
		if err != nil {
			return nil, s[offset+olen:], fmt.Errorf("cannot parse object: %s", err)
		}
		return v, s[offset+olen:], nil
	}
	if s[offset] == '[' {
		v, alen, err := parseArray(s, offset, c, depth)
		if err != nil {
			return nil, s[offset+alen:], fmt.Errorf("cannot parse array: %s", err)
		}
		return v, s[offset+alen:], nil
	}
	if s[offset] == '"' {
		ss, slen, err := parseRawString(s, offset)
		if err != nil {
			return nil, s[offset+slen:], fmt.Errorf("cannot parse string: %s", err)
		}
		v := c.getValue()
		v.t = typeRawString
		v.s = ss
		v.do = offset
		v.dl = slen
		return v, s[offset+slen:], nil
	}
	if s[offset] == 't' {
		if len(s[offset:]) < len("true") || s[offset:offset + len("true")] != "true" {
			return nil, s, fmt.Errorf("unexpected value found: %q", s[offset:])
		}
		v := c.getValue()
		v.t = valueTrue.t
		v.do = offset
		v.dl = valueTrue.dl
		return v, s[offset+v.dl:], nil
	}
	if s[offset] == 'f' {
		if len(s[offset:]) < len("false") || s[offset:offset + len("false")] != "false" {
			return nil, s, fmt.Errorf("unexpected value found: %q", s[offset:])
		}
		v := c.getValue()
		v.t = valueFalse.t
		v.do = offset
		v.dl = valueFalse.dl
		return v, s[offset+v.dl:], nil
	}
	if s[offset] == 'n' {
		if len(s[offset:]) < len("null") || s[offset:offset + len("null")] != "null" {
			// Try parsing NaN
			if len(s[offset:]) >= 3 && strings.EqualFold(s[offset:offset+3], "nan") {
				v := c.getValue()
				v.t = TypeNumber
				v.s = s[offset:offset+3]
				v.do = offset
				v.dl = 3
				return v, s[offset+3:], nil
			}
			return nil, s, fmt.Errorf("unexpected value found: %q", s[offset:])
		}
		v := c.getValue()
		v.t = valueNull.t
		v.do = offset
		v.dl = valueNull.dl
		return v, s[offset+v.dl:], nil
	}

	ns, nlen, err := parseRawNumber(s, offset)
	offset += nlen
	if err != nil {
		return nil, s[offset:], fmt.Errorf("cannot parse number: %s", err)
	}
	v := c.getValue()
	v.t = TypeNumber
	v.s = ns
	v.do = offset
	v.dl = nlen
	return v, s[offset:], nil
}

func parseArray(s string, offset int, c *cache, depth int) (*Value, int, error) {
	start_offset := offset
	offset++
	offset += skipWS(s[offset:])
	if offset >= len(s) {
		return nil, offset - start_offset, fmt.Errorf("missing ']'")
	}

	if s[offset] == ']' {
		v := c.getValue()
		v.t = TypeArray
		v.a = v.a[:0]
		v.do = start_offset
		v.dl = offset - start_offset + 1
		return v, offset - start_offset + 1, nil
	}

	a := c.getValue()
	a.t = TypeArray
	a.a = a.a[:0]
	a.do = start_offset
	for {
		var v *Value
		var err error

		offset += skipWS(s[offset:])
		v, _, err = parseValue(s, offset, c, depth)
		if err != nil {
			return nil, offset, fmt.Errorf("cannot parse array value: %s", err)
		}
		a.a = append(a.a, v)

		offset += v.dl
		offset += skipWS(s[offset:])
		if offset >= len(s) {
			return nil, offset, fmt.Errorf("unexpected end of array")
		}
		if s[offset] == ',' {
			offset++
			continue
		}
		if s[offset] == ']' {
			offset++
			a.dl = offset - start_offset
			return a, a.dl, nil
		}
		return nil, offset - start_offset, fmt.Errorf("missing ',' after array value")
	}
}

func parseObject(s string, offset int, c *cache, depth int) (*Value, int, error) {
	start_offset := offset
	offset++
	offset += skipWS(s[offset:])
	if offset >= len(s) {
		return nil, offset - start_offset, fmt.Errorf("missing '}'")
	}

	if s[offset] == '}' {
		v := c.getValue()
		v.t = TypeObject
		v.o.reset()
		v.do = start_offset
		v.dl = offset - start_offset + 1
		return v, offset - start_offset + 1, nil
	}

	o := c.getValue()
	o.t = TypeObject
	o.o.reset()
	o.do = start_offset
	for {
		var err error
		var klen int
		kv := o.o.getKV()

		// Parse key.
		offset += skipWS(s[offset:])
		if len(s[offset:]) == 0 || s[offset] != '"' {
			return nil, offset - start_offset, fmt.Errorf(`cannot find opening '"" for object key`)
		}
		kv.k, klen, err = parseRawKey(s, offset)
		offset += klen
		if err != nil {
			return nil, offset - start_offset, fmt.Errorf("cannot parse object key: %s", err)
		}
		offset += skipWS(s[offset:])
		if offset >= len(s) || s[offset] != ':' {
			return nil, offset - start_offset, fmt.Errorf("missing ':' after object key")
		}
		offset++

		// Parse value
		offset += skipWS(s[offset:])
		kv.v, _, err = parseValue(s, offset, c, depth)
		if err != nil {
			return nil, offset - start_offset, fmt.Errorf("cannot parse object value: %s", err)
		}
		offset += kv.v.dl
		offset += skipWS(s[offset:])
		if offset >= len(s) {
			return nil, offset - start_offset, fmt.Errorf("unexpected end of object")
		}
		if s[offset] == ',' {
			offset++
			continue
		}
		if s[offset] == '}' {
			o.dl = offset + 1 - start_offset
			return o, o.dl, nil
		}
		return nil, offset - start_offset, fmt.Errorf("missing ',' after object value")
	}
}

func escapeString(dst []byte, s string) []byte {
	if !hasSpecialChars(s) {
		// Fast path - nothing to escape.
		dst = append(dst, '"')
		dst = append(dst, s...)
		dst = append(dst, '"')
		return dst
	}

	// Slow path.
	return strconv.AppendQuote(dst, s)
}

func hasSpecialChars(s string) bool {
	if strings.IndexByte(s, '"') >= 0 || strings.IndexByte(s, '\\') >= 0 {
		return true
	}
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 {
			return true
		}
	}
	return false
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
				// Too short escape sequence. Just store it unchanged.
				b = append(b, "\\u"...)
				break
			}
			xs := s[:4]
			x, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				// Invalid escape sequence. Just store it unchanged.
				b = append(b, "\\u"...)
				break
			}
			s = s[4:]
			if !utf16.IsSurrogate(rune(x)) {
				b = append(b, string(rune(x))...)
				break
			}

			// Surrogate.
			// See https://en.wikipedia.org/wiki/Universal_Character_Set_characters#Surrogates
			if len(s) < 6 || s[0] != '\\' || s[1] != 'u' {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			x1, err := strconv.ParseUint(s[2:6], 16, 16)
			if err != nil {
				b = append(b, "\\u"...)
				b = append(b, xs...)
				break
			}
			r := utf16.DecodeRune(rune(x), rune(x1))
			b = append(b, string(r)...)
			s = s[6:]
		default:
			// Unknown escape sequence. Just store it unchanged.
			b = append(b, '\\', ch)
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

// parseRawKey is similar to parseRawString, but is optimized
// for small-sized keys without escape sequences.
func parseRawKey(s string, offset int) (string, int, error) {
	start_offset := offset
	for i := offset+1; i < len(s); i++ {
		if s[i] == '"' {
			// Fast path.
			return s[start_offset+1:i], i-offset+1 /* include quotes */, nil
		}
		if s[i] == '\\' {
			// Slow path.
			return parseRawString(s, start_offset)
		}
	}
	return s, len(s[start_offset:]), fmt.Errorf(`missing closing '"'`)
}

func parseRawString(s string, offset int) (string, int, error) {
	start_offset := offset
	offset++
	if offset >= len(s) {
		return "", 1, fmt.Errorf(`missing closing '"'`)
	}
	n := strings.IndexByte(s[offset:], '"')
	if n < 0 {
		return "", len(s[start_offset:]), fmt.Errorf(`missing closing '"'`)
	}
	if n == 0 || s[offset+n-1] != '\\' {
		// Fast path. No escaped ".
		return s[offset:offset+n], n+1+1 /* include quotes */, nil
	}

	// Slow path - possible escaped " found.
	ss := s[offset:]
	for {
		i := n - 1
		for i > 0 && s[offset+i-1] == '\\' {
			i--
		}
		if uint(n-i)%2 == 0 {
			return ss[:len(ss)-len(s[offset:])+n], offset-start_offset+n+1, nil
		}
		offset += n+1

		n = strings.IndexByte(s[offset:], '"')
		if n < 0 {
			return ss, len(s[start_offset:]), fmt.Errorf(`missing closing '"'`)
		}
		if n == 0 || s[offset+n-1] != '\\' {
			return ss[:len(ss)-len(s[offset:])+n], offset-start_offset+n+1, nil
		}
	}
}

func parseRawNumber(s string, offset int) (string, int, error) {
	// The caller must ensure len(s[offset:]) > 0
	start_offset := offset

	// Find the end of the number.
	for i := 0; start_offset+i < len(s); i++ {
		ch := s[offset]
		if (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' || ch == 'e' || ch == 'E' || ch == '+' {
			offset++
			continue
		}
		if i == 0 || i == 1 && (s[offset-1] == '-' || s[offset-1] == '+') {
			if len(s[offset:]) >= 3 {
				// offset += i
				xs := s[offset : offset+3]
				if strings.EqualFold(xs, "inf") || strings.EqualFold(xs, "nan") {
					return s[start_offset:offset+3], offset-start_offset+3, nil
				}
			}
			return "", 0, fmt.Errorf("unexpected char: %q", s[offset:offset+1])
		}
		ns := s[start_offset:offset]
		return ns, offset - start_offset, nil
	}
	return s, offset - start_offset, nil
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

// MarshalTo appends marshaled o to dst and returns the result.
func (o *Object) MarshalTo(dst []byte) []byte {
	dst = append(dst, '{')
	for i, kv := range o.kvs {
		if o.keysUnescaped {
			dst = escapeString(dst, kv.k)
		} else {
			dst = append(dst, '"')
			dst = append(dst, kv.k...)
			dst = append(dst, '"')
		}
		dst = append(dst, ':')
		dst = kv.v.MarshalTo(dst)
		if i != len(o.kvs)-1 {
			dst = append(dst, ',')
		}
	}
	dst = append(dst, '}')
	return dst
}

// String returns string representation for the o.
//
// This function is for debugging purposes only. It isn't optimized for speed.
// See MarshalTo instead.
func (o *Object) String() string {
	b := o.MarshalTo(nil)
	// It is safe converting b to string without allocation, since b is no longer
	// reachable after this line.
	return b2s(b)
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
	kvs := o.kvs
	for i := range kvs {
		kv := &kvs[i]
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
	if !o.keysUnescaped && strings.IndexByte(key, '\\') < 0 {
		// Fast path - try searching for the key without object keys unescaping.
		for _, kv := range o.kvs {
			if kv.k == key {
				return kv.v
			}
		}
	}

	// Slow path - unescape object keys.
	o.unescapeKeys()

	for _, kv := range o.kvs {
		if kv.k == key {
			return kv.v
		}
	}
	return nil
}

// Visit calls f for each item in the o in the original order
// of the parsed JSON.
//
// f cannot hold key and/or v after returning.
func (o *Object) Visit(f func(key []byte, v *Value)) {
	if o == nil {
		return
	}

	o.unescapeKeys()

	for _, kv := range o.kvs {
		f(s2b(kv.k), kv.v)
	}
}

// Value represents any JSON value.
//
// Call Type in order to determine the actual type of the JSON value.
//
// Value cannot be used from concurrent goroutines.
// Use per-goroutine parsers or ParserPool instead.
type Value struct {
	o Object
	a []*Value
	s string
	t Type
	do int
	dl int
}

// MarshalTo appends marshaled v to dst and returns the result.
func (v *Value) MarshalTo(dst []byte) []byte {
	switch v.t {
	case typeRawString:
		dst = append(dst, '"')
		dst = append(dst, v.s...)
		dst = append(dst, '"')
		return dst
	case TypeObject:
		return v.o.MarshalTo(dst)
	case TypeArray:
		dst = append(dst, '[')
		for i, vv := range v.a {
			dst = vv.MarshalTo(dst)
			if i != len(v.a)-1 {
				dst = append(dst, ',')
			}
		}
		dst = append(dst, ']')
		return dst
	case TypeString:
		return escapeString(dst, v.s)
	case TypeNumber:
		return append(dst, v.s...)
	case TypeTrue:
		return append(dst, "true"...)
	case TypeFalse:
		return append(dst, "false"...)
	case TypeNull:
		return append(dst, "null"...)
	default:
		panic(fmt.Errorf("BUG: unexpected Value type: %d", v.t))
	}
}

// String returns string representation of the v.
//
// The function is for debugging purposes only. It isn't optimized for speed.
// See MarshalTo instead.
//
// Don't confuse this function with StringBytes, which must be called
// for obtaining the underlying JSON string for the v.
func (v *Value) String() string {
	b := v.MarshalTo(nil)
	// It is safe converting b to string without allocation, since b is no longer
	// reachable after this line.
	return b2s(b)
}

// Type represents JSON type.
type Type int

const (
	// TypeNull is JSON null.
	TypeNull Type = 0

	// TypeObject is JSON object type.
	TypeObject Type = 1

	// TypeArray is JSON array type.
	TypeArray Type = 2

	// TypeString is JSON string type.
	TypeString Type = 3

	// TypeNumber is JSON number type.
	TypeNumber Type = 4

	// TypeTrue is JSON true.
	TypeTrue Type = 5

	// TypeFalse is JSON false.
	TypeFalse Type = 6

	typeRawString Type = 7
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

	// typeRawString is skipped intentionally,
	// since it shouldn't be visible to user.
	default:
		panic(fmt.Errorf("BUG: unknown Value type: %d", t))
	}
}

// Type returns the type of the v.
func (v *Value) Type() Type {
	if v.t == typeRawString {
		v.s = unescapeStringBestEffort(v.s)
		v.t = TypeString
	}
	return v.t
}

// Offset returns the zero-indexed offset of the v in the original JSON string.
func (v *Value) Offset() int {
	if v == nil {
		return 0
	}
	return v.do
}

// Len returns the length of the v in the original JSON string.
func (v *Value) Len() int {
	if v == nil {
		return 0
	}
	return v.dl
}

// Exists returns true if the field exists for the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
func (v *Value) Exists(keys ...string) bool {
	v = v.Get(keys...)
	return v != nil
}

// Get returns value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned for non-existing keys path.
//
// The returned value is valid until Parse is called on the Parser returned v.
func (v *Value) Get(keys ...string) *Value {
	if v == nil {
		return nil
	}
	for _, key := range keys {
		if v.t == TypeObject {
			v = v.o.Get(key)
			if v == nil {
				return nil
			}
		} else if v.t == TypeArray {
			n, err := strconv.Atoi(key)
			if err != nil || n < 0 || n >= len(v.a) {
				return nil
			}
			v = v.a[n]
		} else {
			return nil
		}
	}
	return v
}

// GetObject returns object value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned for non-existing keys path or for invalid value type.
//
// The returned object is valid until Parse is called on the Parser returned v.
func (v *Value) GetObject(keys ...string) *Object {
	v = v.Get(keys...)
	if v == nil || v.t != TypeObject {
		return nil
	}
	return &v.o
}

// GetArray returns array value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned for non-existing keys path or for invalid value type.
//
// The returned array is valid until Parse is called on the Parser returned v.
func (v *Value) GetArray(keys ...string) []*Value {
	v = v.Get(keys...)
	if v == nil || v.t != TypeArray {
		return nil
	}
	return v.a
}

// GetFloat64 returns float64 value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetFloat64(keys ...string) float64 {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	return fastfloat.ParseBestEffort(v.s)
}

// GetInt returns int value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetInt(keys ...string) int {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	n := fastfloat.ParseInt64BestEffort(v.s)
	nn := int(n)
	if int64(nn) != n {
		return 0
	}
	return nn
}

// GetUint returns uint value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetUint(keys ...string) uint {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	n := fastfloat.ParseUint64BestEffort(v.s)
	nn := uint(n)
	if uint64(nn) != n {
		return 0
	}
	return nn
}

// GetInt64 returns int64 value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetInt64(keys ...string) int64 {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	return fastfloat.ParseInt64BestEffort(v.s)
}

// GetUint64 returns uint64 value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned for non-existing keys path or for invalid value type.
func (v *Value) GetUint64(keys ...string) uint64 {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeNumber {
		return 0
	}
	return fastfloat.ParseUint64BestEffort(v.s)
}

// GetStringBytes returns string value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned for non-existing keys path or for invalid value type.
//
// The returned string is valid until Parse is called on the Parser returned v.
func (v *Value) GetStringBytes(keys ...string) []byte {
	v = v.Get(keys...)
	if v == nil || v.Type() != TypeString {
		return nil
	}
	return s2b(v.s)
}

// GetBool returns bool value by the given keys path.
//
// Array indexes may be represented as decimal numbers in keys.
//
// false is returned for non-existing keys path or for invalid value type.
func (v *Value) GetBool(keys ...string) bool {
	v = v.Get(keys...)
	if v != nil && v.t == TypeTrue {
		return true
	}
	return false
}

// Object returns the underlying JSON object for the v.
//
// The returned object is valid until Parse is called on the Parser returned v.
//
// Use GetObject if you don't need error handling.
func (v *Value) Object() (*Object, error) {
	if v.t != TypeObject {
		return nil, fmt.Errorf("value doesn't contain object; it contains %s", v.Type())
	}
	return &v.o, nil
}

// Array returns the underlying JSON array for the v.
//
// The returned array is valid until Parse is called on the Parser returned v.
//
// Use GetArray if you don't need error handling.
func (v *Value) Array() ([]*Value, error) {
	if v.t != TypeArray {
		return nil, fmt.Errorf("value doesn't contain array; it contains %s", v.Type())
	}
	return v.a, nil
}

// StringBytes returns the underlying JSON string for the v.
//
// The returned string is valid until Parse is called on the Parser returned v.
//
// Use GetStringBytes if you don't need error handling.
func (v *Value) StringBytes() ([]byte, error) {
	if v.Type() != TypeString {
		return nil, fmt.Errorf("value doesn't contain string; it contains %s", v.Type())
	}
	return s2b(v.s), nil
}

// Float64 returns the underlying JSON number for the v.
//
// Use GetFloat64 if you don't need error handling.
func (v *Value) Float64() (float64, error) {
	if v.Type() != TypeNumber {
		return 0, fmt.Errorf("value doesn't contain number; it contains %s", v.Type())
	}
	return fastfloat.Parse(v.s)
}

// Int returns the underlying JSON int for the v.
//
// Use GetInt if you don't need error handling.
func (v *Value) Int() (int, error) {
	if v.Type() != TypeNumber {
		return 0, fmt.Errorf("value doesn't contain number; it contains %s", v.Type())
	}
	n, err := fastfloat.ParseInt64(v.s)
	if err != nil {
		return 0, err
	}
	nn := int(n)
	if int64(nn) != n {
		return 0, fmt.Errorf("number %q doesn't fit int", v.s)
	}
	return nn, nil
}

// Uint returns the underlying JSON uint for the v.
//
// Use GetInt if you don't need error handling.
func (v *Value) Uint() (uint, error) {
	if v.Type() != TypeNumber {
		return 0, fmt.Errorf("value doesn't contain number; it contains %s", v.Type())
	}
	n, err := fastfloat.ParseUint64(v.s)
	if err != nil {
		return 0, err
	}
	nn := uint(n)
	if uint64(nn) != n {
		return 0, fmt.Errorf("number %q doesn't fit uint", v.s)
	}
	return nn, nil
}

// Int64 returns the underlying JSON int64 for the v.
//
// Use GetInt64 if you don't need error handling.
func (v *Value) Int64() (int64, error) {
	if v.Type() != TypeNumber {
		return 0, fmt.Errorf("value doesn't contain number; it contains %s", v.Type())
	}
	return fastfloat.ParseInt64(v.s)
}

// Uint64 returns the underlying JSON uint64 for the v.
//
// Use GetInt64 if you don't need error handling.
func (v *Value) Uint64() (uint64, error) {
	if v.Type() != TypeNumber {
		return 0, fmt.Errorf("value doesn't contain number; it contains %s", v.Type())
	}
	return fastfloat.ParseUint64(v.s)
}

// Bool returns the underlying JSON bool for the v.
//
// Use GetBool if you don't need error handling.
func (v *Value) Bool() (bool, error) {
	if v.t == TypeTrue {
		return true, nil
	}
	if v.t == TypeFalse {
		return false, nil
	}
	return false, fmt.Errorf("value doesn't contain bool; it contains %s", v.Type())
}

var (
	valueTrue  = &Value{t: TypeTrue, dl: 4}
	valueFalse = &Value{t: TypeFalse, dl: 5}
	valueNull  = &Value{t: TypeNull, dl: 4}
)
