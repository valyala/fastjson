package fastjson

import (
	"fmt"
	"strconv"
	"strings"
)

// ValidParser validates while parsing JSON.
//
// ValidParser may be re-used for subsequent parsing.
//
// ValidParser cannot be used from concurrent goroutines.
// Use per-goroutine ValidParsers or ValidParserPool instead.
type ValidParser struct {
	// b contains working copy of the string to be parsed.
	b []byte

	// c is a cache for json values.
	c cache
}

// Parse parses and validates s containing JSON.
//
// The returned value is valid until the next call to Parse*.
//
// Use Scanner if a stream of JSON values must be parsed and validated.
func (p *ValidParser) Parse(s string) (*Value, error) {
	s = skipWS(s)
	p.b = append(p.b[:0], s...)
	p.c.reset()

	v, tail, err := parseValidValue(b2s(p.b), &p.c, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot parseValid JSON: %s; unparsed tail: %q", err, startEndString(tail))
	}
	tail = skipWS(tail)
	if len(tail) > 0 {
		return nil, fmt.Errorf("unexpected tail: %q", startEndString(tail))
	}
	return v, nil
}

// ParseBytes parses and validates b containing JSON.
//
// The returned Value is valid until the next call to Parse*.
//
// Use Scanner if a stream of JSON values must be parsed and validated.
func (p *ValidParser) ParseBytes(b []byte) (*Value, error) {
	return p.Parse(b2s(b))
}

func parseValidValue(s string, c *cache, depth int) (*Value, string, error) {
	if len(s) == 0 {
		return nil, s, fmt.Errorf("cannot parseValid empty string")
	}
	depth++
	if depth > MaxDepth {
		return nil, s, fmt.Errorf("too big depth for the nested JSON; it exceeds %d", MaxDepth)
	}

	if s[0] == '{' {
		v, tail, err := parseValidObject(s[1:], c, depth)
		if err != nil {
			return nil, tail, fmt.Errorf("cannot parseValid object: %s", err)
		}
		return v, tail, nil
	}
	if s[0] == '[' {
		v, tail, err := parseValidArray(s[1:], c, depth)
		if err != nil {
			return nil, tail, fmt.Errorf("cannot parseValid array: %s", err)
		}
		return v, tail, nil
	}
	if s[0] == '"' {
		ss, tail, err := parseValidRawString(s[1:])
		if err != nil {
			return nil, tail, fmt.Errorf("cannot parseValid string: %s", err)
		}
		// Scan the string for control chars.
		for i := 0; i < len(ss); i++ {
			if ss[i] < 0x20 {
				return nil, tail, fmt.Errorf("string cannot contain control char 0x%02X", ss[i])
			}
		}
		v := c.getValue()
		v.t = typeRawString
		v.s = ss
		return v, tail, nil
	}
	if s[0] == 't' {
		if len(s) < len("true") || s[:len("true")] != "true" {
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		return valueTrue, s[len("true"):], nil
	}
	if s[0] == 'f' {
		if len(s) < len("false") || s[:len("false")] != "false" {
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		return valueFalse, s[len("false"):], nil
	}
	if s[0] == 'n' {
		if len(s) < len("null") || s[:len("null")] != "null" {
			return nil, s, fmt.Errorf("unexpected value found: %q", s)
		}
		return valueNull, s[len("null"):], nil
	}

	ns, tail, err := parseValidRawNumber(s)
	if err != nil {
		return nil, tail, fmt.Errorf("cannot parseValid number: %s", err)
	}
	v := c.getValue()
	v.t = TypeNumber
	v.s = ns
	return v, tail, nil
}

func parseValidArray(s string, c *cache, depth int) (*Value, string, error) {
	s = skipWS(s)
	if len(s) == 0 {
		return nil, s, fmt.Errorf("missing ']'")
	}

	if s[0] == ']' {
		v := c.getValue()
		v.t = TypeArray
		v.a = v.a[:0]
		return v, s[1:], nil
	}

	a := c.getValue()
	a.t = TypeArray
	a.a = a.a[:0]
	for {
		var v *Value
		var err error

		s = skipWS(s)
		v, s, err = parseValidValue(s, c, depth)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parseValid array value: %s", err)
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

func parseValidObject(s string, c *cache, depth int) (*Value, string, error) {
	s = skipWS(s)
	if len(s) == 0 {
		return nil, s, fmt.Errorf("missing '}'")
	}

	if s[0] == '}' {
		v := c.getValue()
		v.t = TypeObject
		v.o.reset()
		return v, s[1:], nil
	}

	o := c.getValue()
	o.t = TypeObject
	o.o.reset()
	for {
		var err error
		kv := o.o.getKV()

		// Parse key.
		s = skipWS(s)
		if len(s) == 0 || s[0] != '"' {
			return nil, s, fmt.Errorf(`cannot find opening '"" for object key`)
		}
		kv.k, s, err = parseValidRawKey(s[1:])
		if err != nil {
			return nil, s, fmt.Errorf("cannot parseValid object key: %s", err)
		}
		s = skipWS(s)
		if len(s) == 0 || s[0] != ':' {
			return nil, s, fmt.Errorf("missing ':' after object key")
		}
		s = s[1:]

		// Parse value
		s = skipWS(s)
		kv.v, s, err = parseValidValue(s, c, depth)
		if err != nil {
			return nil, s, fmt.Errorf("cannot parseValid object value: %s", err)
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

// parseValidRawKey is similar to parseValidRawString, but is optimized
// for small-sized keys without escape sequences.
func parseValidRawKey(s string) (string, string, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			// Fast path.
			return s[:i], s[i+1:], nil
		}
		if s[i] == '\\' {
			// Slow path.
			return parseValidRawString(s)
		}
	}
	return s, "", fmt.Errorf(`missing closing '"'`)
}

func parseValidRawString(s string) (string, string, error) {
	// Try fast path - a string without escape sequences.
	if n := strings.IndexByte(s, '"'); n >= 0 && strings.IndexByte(s[:n], '\\') < 0 {
		return s[:n], s[n+1:], nil
	}

	// Slow path - escape sequences are present.
	prs, tail, err := parseRawString(s)
	if err != nil {
		return prs, tail, err
	}
	var rs = prs
	for {
		n := strings.IndexByte(rs, '\\')
		if n < 0 {
			return prs, tail, nil
		}
		n++
		ch := rs[n]
		rs = rs[n+1:]
		switch ch {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			// Valid escape sequences - see http://json.org/
			break
		case 'u':
			if len(rs) < 4 {
				return prs, tail, fmt.Errorf(`too short escape sequence: \u%s`, rs)
			}
			xs := rs[:4]
			_, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				return prs, tail, fmt.Errorf(`invalid escape sequence \u%s: %s`, xs, err)
			}
			rs = rs[4:]
		default:
			return prs, tail, fmt.Errorf(`unknown escape sequence \%c`, ch)
		}
	}
}

func parseValidRawNumber(s string) (string, string, error) {
	if len(s) == 0 {
		return "", s, fmt.Errorf("zero-length number")
	}
	i := 0
	/*
		 * Validator does not Support Inf/NaN. Parser does.
		 * Choosing not to support it in ValidParser in order to match JSON spec and behavior of encoding/json.
		 *
		if len(s[i:]) >= 3 {
			xs := s[i : i+3]
			if strings.EqualFold(xs, "inf") || strings.EqualFold(xs, "nan") {
				return s[:i+3], s[i+3:], nil
			}
		}
	*/
	if s[0] == '-' {
		i++
		if len(s) == i {
			return "", s, fmt.Errorf("missing number after minus")
		}
	}
	var j = i
	for i < len(s) {
		if s[i] < '0' || s[i] > '9' {
			break
		}
		i++
	}
	if j == i {
		return "", s, fmt.Errorf("expecting 0..9 digit, got %c", s[0])
	}
	if s[j] == '0' && i-j != 1 {
		return "", s, fmt.Errorf("unexpected number starting from 0")
	}
	if i >= len(s) {
		return s[:i], s[i:], nil
	}
	if s[i] == '.' {
		// Validate fractional part
		i++
		if len(s) == i {
			return "", s, fmt.Errorf("missing fractional part")
		}
		j = i
		for i < len(s) {
			if s[i] < '0' || s[i] > '9' {
				break
			}
			i++
		}
		if j == i {
			return "", s, fmt.Errorf("expecting 0..9 digit in exponent part, got %c", s[i])
		}
		if len(s) == i {
			return s[:i], s[i:], nil
		}
	}
	if s[i] == 'e' || s[i] == 'E' {
		// Validate exponent part
		i++
		if len(s) == i {
			return "", s, fmt.Errorf("missing exponent part")
		}
		if s[i] == '-' || s[i] == '+' {
			i++
			if len(s) == i {
				return "", s, fmt.Errorf("missing exponent part")
			}
		}
		j = i
		for i < len(s) {
			if s[i] < '0' || s[i] > '9' {
				break
			}
			i++
		}
		if j == i {
			return "", s, fmt.Errorf("expecting 0..9 digit in exponent part, got %c", s[i])
		}
	}
	return s[:i], s[i:], nil
}
