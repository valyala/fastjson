package fastjson

import (
	"fmt"
)

// Validate validates JSON s.
func Validate(s string) error {
	s = skipWS(s)

	tail, err := validateValue(s)
	if err != nil {
		return fmt.Errorf("cannot parse JSON: %s; unparsed tail: %q", err, startEndString(tail))
	}
	tail = skipWS(tail)
	if len(tail) > 0 {
		return fmt.Errorf("unexpected tail: %q", startEndString(tail))
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

	switch s[0] {
	case '{':
		tail, err := validateObject(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse object: %s", err)
		}
		return tail, nil
	case '[':
		tail, err := validateArray(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse array: %s", err)
		}
		return tail, nil
	case '"':
		_, tail, err := validateString(s[1:])
		if err != nil {
			return tail, fmt.Errorf("cannot parse string: %s", err)
		}
		return tail, nil
	case 't':
		if len(s) < len("true") || s[:len("true")] != "true" {
			return s, fmt.Errorf("unexpected value found: %q", s)
		}
		return s[len("true"):], nil
	case 'f':
		if len(s) < len("false") || s[:len("false")] != "false" {
			return s, fmt.Errorf("unexpected value found: %q", s)
		}
		return s[len("false"):], nil
	case 'n':
		if len(s) < len("null") || s[:len("null")] != "null" {
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
			return s[1:], nil
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

		_, s, err = validateKey(s[1:])
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

// validateKey is similar to validateString, but is optimized
// for typical object keys, which are quite small and have no escape sequences.
func validateKey(s string) (string, string, error) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			// Fast path - the key doesn't contain escape sequences.
			return s[:i], s[i+1:], nil
		}
		if c == '\\' {
			// Slow path - the key contains escape sequences.
			return validateStringSlow(s, i)
		}
		if c < 0x20 {
			return s[:i], s[i:], fmt.Errorf("object key cannot contain control char 0x%02X", c)
		}
	}
	return "", s, fmt.Errorf(`missing closing '"'`)
}

func validateString(s string) (string, string, error) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			return s[:i], s[i+1:], nil
		}
		if c == '\\' {
			return validateStringSlow(s, i)
		}
		if c < 0x20 {
			return s[:i], s[i:], fmt.Errorf("string cannot contain control char 0x%02X", c)
		}
	}
	return "", s, fmt.Errorf(`missing closing '"'`)
}

func validateStringSlow(s string, i int) (string, string, error) {
	for ; i < len(s); i++ {
		c := s[i]
		if c == '"' {
			return s[:i], s[i+1:], nil
		}
		if c < 0x20 {
			return s[:i], s[i:], fmt.Errorf("string cannot contain control char 0x%02X", c)
		}
		if c != '\\' {
			continue
		}

		i++
		if i >= len(s) {
			return s, "", fmt.Errorf(`missing closing '"'`)
		}
		ch := s[i]
		switch ch {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			// Valid escape sequences - see http://json.org/
		case 'u':
			if i+4 >= len(s) {
				return s[i+1:], "", fmt.Errorf(`too short escape sequence: \u%s`, s[i+1:])
			}
			for j := i + 1; j <= i+4; j++ {
				c := s[j]
				if !isHexChar(c) {
					return s[j:], "", fmt.Errorf(`invalid escape sequence \u%s`, s[i+1:i+5])
				}
			}
			i += 4
		default:
			return s[i:], "", fmt.Errorf(`unknown escape sequence \%c`, ch)
		}
	}
	return "", s, fmt.Errorf(`missing closing '"'`)
}

func isHexChar(c byte) bool {
	return ('0' <= c && c <= '9') || ('a' <= c && c <= 'f') || ('A' <= c && c <= 'F')
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
	if i <= 0 {
		return s, fmt.Errorf("expecting 0..9 digit, got %c", s[0])
	}
	if s[0] == '0' && i != 1 {
		return s, fmt.Errorf("unexpected number starting from 0")
	}
	if i >= len(s) {
		return "", nil
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
