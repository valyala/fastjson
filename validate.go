package fastjson

import (
	"fmt"
	"strconv"
	"strings"
)

// Validate validates JSON s.
func Validate(s string) error {
	s = skipWS(s)

	tail, err := validateValue(s)
	if err != nil {
		return fmt.Errorf("cannot parse JSON: %s; unparsed tail: %q", err, tail)
	}
	tail = skipWS(tail)
	if len(tail) > 0 {
		return fmt.Errorf("unexpected tail: %q", tail)
	}
	return nil
}

// ValidateBytes validates JSON b.
func ValidateBytes(b []byte) error {
	return Validate(b2s(b))
}

func validateValue(s string) (string, error) {
	if len(s) == 0 {
		return s, fmt.Errorf("cannot parse empty string")
	}

	if s[0] == '{' {
		tail, err := validateObject(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse object: %s", err)
		}
		return tail, nil
	}
	if s[0] == '[' {
		tail, err := validateArray(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse array: %s", err)
		}
		return tail, nil
	}
	if s[0] == '"' {
		tail, err := validateString(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse string: %s", err)
		}
		return tail, nil
	}
	if s[0] == 't' {
		if !strings.HasPrefix(s, "true") {
			return s, fmt.Errorf("unexpected value found: %q", s)
		}
		return s[len("true"):], nil
	}
	if s[0] == 'f' {
		if !strings.HasPrefix(s, "false") {
			return s, fmt.Errorf("unexpected value found: %q", s)
		}
		return s[len("false"):], nil
	}
	if s[0] == 'n' {
		if !strings.HasPrefix(s, "null") {
			return s, fmt.Errorf("unexpected value found: %q", s)
		}
		return s[len("null"):], nil
	}

	tail, err := validateNumber(s)
	if err != nil {
		return tail, fmt.Errorf("cannot parse number: %s", err)
	}
	return tail, nil
}

func validateArray(s string) (string, error) {
	s = skipWS(s)
	if len(s) == 0 {
		return s, fmt.Errorf("missing ']'")
	}
	if s[0] == ']' {
		return s[1:], nil
	}

	for {
		var err error

		s = skipWS(s)
		s, err = validateValue(s)
		if err != nil {
			return s, fmt.Errorf("cannot parse array value: %s", err)
		}

		s = skipWS(s)
		if len(s) == 0 {
			return s, fmt.Errorf("unexpected end of array")
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		if s[0] == ']' {
			s = s[1:]
			return s, nil
		}
		return s, fmt.Errorf("missing ',' after array value")
	}
}

func validateObject(s string) (string, error) {
	s = skipWS(s)
	if len(s) == 0 {
		return s, fmt.Errorf("missing '}'")
	}
	if s[0] == '}' {
		return s[1:], nil
	}

	for {
		var err error

		// Parse key.
		s = skipWS(s)
		if len(s) == 0 || s[0] != '"' {
			return s, fmt.Errorf(`cannot find opening '"" for object key`)
		}
		s, err = validateString(s[1:])
		if err != nil {
			return s, fmt.Errorf("cannot parse object key: %s", err)
		}
		s = skipWS(s)
		if len(s) == 0 || s[0] != ':' {
			return s, fmt.Errorf("missing ':' after object key")
		}
		s = s[1:]

		// Parse value
		s = skipWS(s)
		s, err = validateValue(s)
		if err != nil {
			return s, fmt.Errorf("cannot parse object value: %s", err)
		}
		s = skipWS(s)
		if len(s) == 0 {
			return s, fmt.Errorf("unexpected end of object")
		}
		if s[0] == ',' {
			s = s[1:]
			continue
		}
		if s[0] == '}' {
			return s[1:], nil
		}
		return s, fmt.Errorf("missing ',' after object value")
	}
}

func validateString(s string) (string, error) {
	rs, tail, err := parseRawString(s)
	if err != nil {
		return tail, err
	}
	n := strings.IndexByte(rs, '\\')
	if n < 0 {
		// Fast path - no escape chars.
		return tail, nil
	}

	// Slow path - escape chars present.
	rs = rs[n+1:]
	for len(rs) > 0 {
		ch := rs[0]
		rs = rs[1:]
		switch ch {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			// Valid escape sequences - see http://json.org/
			break
		case 'u':
			if len(rs) < 4 {
				return tail, fmt.Errorf(`too short escape sequence: \u%s`, rs)
			}
			xs := rs[:4]
			_, err := strconv.ParseUint(xs, 16, 16)
			if err != nil {
				return tail, fmt.Errorf(`invalid escape sequence \u%s: %s`, xs, err)
			}
			rs = rs[4:]
		default:
			return tail, fmt.Errorf(`unknown escape sequence \%c`, ch)
		}
		n = strings.IndexByte(rs, '\\')
		if n < 0 {
			break
		}
		rs = rs[n+1:]
	}
	return tail, nil
}

func validateNumber(s string) (string, error) {
	if len(s) == 0 {
		return s, fmt.Errorf("zero-length number")
	}
	if s[0] == '-' {
		s = s[1:]
		if len(s) == 0 {
			return s, fmt.Errorf("missing number after minus")
		}
	}
	i := 0
	for i < len(s) {
		if s[i] < '0' || s[i] > '9' {
			break
		}
		i++
	}
	if i == 0 {
		return s, fmt.Errorf("expecting 0..9 digit, got %c", s[0])
	}
	if i >= len(s) {
		return "", nil
	}
	if s[0] == '0' && i != 1 {
		return s, fmt.Errorf("unexpected number starting from 0")
	}
	if s[i] == '.' {
		// Validate fractional part
		s = s[i+1:]
		if len(s) == 0 {
			return s, fmt.Errorf("missing fractional part")
		}
		i = 0
		for i < len(s) {
			if s[i] < '0' || s[i] > '9' {
				break
			}
			i++
		}
		if i == 0 {
			return s, fmt.Errorf("expecting 0..9 digit in fractional part, got %c", s[0])
		}
		if i >= len(s) {
			return "", nil
		}
	}
	if s[i] == 'e' || s[i] == 'E' {
		// Validate exponent part
		s = s[i+1:]
		if len(s) == 0 {
			return s, fmt.Errorf("missing exponent part")
		}
		if s[0] == '-' || s[0] == '+' {
			s = s[1:]
			if len(s) == 0 {
				return s, fmt.Errorf("missing exponent part")
			}
		}
		i = 0
		for i < len(s) {
			if s[i] < '0' || s[i] > '9' {
				break
			}
			i++
		}
		if i == 0 {
			return s, fmt.Errorf("expecting 0..9 digit in exponent part, got %c", s[0])
		}
		if i >= len(s) {
			return "", nil
		}
	}
	return s[i:], nil
}
