package fastjson

import (
	"encoding/json"
	"testing"
)

func TestEscapeStringValidJSON(t *testing.T) {
	testCases := []struct {
		s    string
		want string
	}{
		{"plain", `"plain"`},
		{"quote \" slash \\", `"quote \" slash \\"`},
		{"\x00\x01\b\f\n\r\t\x0b\x1f", "\"\\u0000\\u0001\\b\\f\\n\\r\\t\\u000b\\u001f\""},
		{"hello \x14 world", "\"hello \\u0014 world\""},
	}
	for _, tc := range testCases {
		dst := escapeString(nil, tc.s)
		if string(dst) != tc.want {
			t.Errorf("escapeString(%q) = %q, want %q", tc.s, dst, tc.want)
		}
		if !json.Valid(dst) {
			t.Errorf("escapeString(%q) produced invalid JSON: %q", tc.s, dst)
		}
	}
}
