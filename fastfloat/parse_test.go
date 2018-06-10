package fastfloat

import (
	"math"
	"testing"
)

func TestParseBestEffort(t *testing.T) {
	f := func(s string, expectedNum float64) {
		t.Helper()

		num := ParseBestEffort(s)
		if num != expectedNum {
			t.Fatalf("unexpected number parsed from %q; got %v; want %v", s, num, expectedNum)
		}
	}

	// Invalid first char
	f("", 0)
	f("  ", 0)
	f("foo", 0)
	f(" bar ", 0)
	f("-", 0)
	f("--", 0)
	f("-.", 0)
	f("-.e", 0)
	f("+112", 0)
	f("++", 0)
	f("e123", 0)
	f("E123", 0)
	f("-e12", 0)
	f(".", 0)
	f("..34", 0)
	f("-.32", 0)
	f("-.e3", 0)
	f(".e+3", 0)

	// Invalid suffix
	f("1foo", 0)
	f("1  foo", 0)
	f("12.34.56", 0)
	f("13e34.56", 0)
	f("12.34e56e4", 0)
	f("12.", 0)
	f("123..45", 0)
	f("123ee34", 0)
	f("123e", 0)
	f("123e+", 0)
	f("123E-", 0)
	f("123E+.", 0)
	f("-123e-23foo", 0)

	// Integer
	f("0", 0)
	f("-0", 0)
	f("0123", 123)
	f("-00123", -123)
	f("1", 1)
	f("-1", -1)
	f("12345678901234567890", 12345678901234567890)

	// Fractional part
	f("0.1", 0.1)
	f("-0.123", -0.123)
	f("12345.12345678901234567890", 12345.12345678901234567890)
	f("-12345.12345678901234567890", -12345.12345678901234567890)

	// Exponent part
	f("0e0", 0)
	f("123e+001", 123e1)
	f("0e12", 0)
	f("-0E123", 0)
	f("-0E-123", 0)
	f("-0E+123", 0)
	f("123e12", 123e12)
	f("-123E-12", -123E-12)
	f("-123e-400", 0)
	f("123e456", math.Inf(1))   // too big exponent
	f("-123e456", math.Inf(-1)) // too big exponent

	// Fractional + exponent part
	f("0.123e4", 0.123e4)
	f("-123.456E-10", -123.456E-10)
}
