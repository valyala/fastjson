package fastjson

import (
	"encoding/json"
	"testing"
)

func TestEscapeStringControlCharsJSONCompatible(t *testing.T) {
	// Control bytes must be emitted as JSON \u00XX (or short escapes),
	// never as Go strconv \xNN which Validate and encoding/json reject.
	cases := []string{
		"hello \x14 world",
		"nul:\x00",
		"bell:\x07",
		"bs:\b tab:\t nl:\n cr:\r ff:\f",
		"quote:\" slash:\\",
	}
	var a Arena
	for _, s := range cases {
		std, err := json.Marshal(s)
		if err != nil {
			t.Fatalf("json.Marshal(%q): %v", s, err)
		}
		fj := a.NewString(s).MarshalTo(nil)
		if err := ValidateBytes(fj); err != nil {
			t.Fatalf("ValidateBytes(%q) failed for input %q: %v (out=%s)", s, s, err, fj)
		}
		var got string
		if err := json.Unmarshal(fj, &got); err != nil {
			t.Fatalf("encoding/json.Unmarshal failed for input %q: %v (out=%s)", s, err, fj)
		}
		if got != s {
			t.Fatalf("round-trip mismatch for %q: got %q", s, got)
		}
		// Match encoding/json for these cases (no HTML meta chars).
		if string(fj) != string(std) {
			t.Fatalf("marshal mismatch for %q:\n  fastjson=%s\n  stdjson =%s", s, fj, std)
		}
	}
}
