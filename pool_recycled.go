package fastjson

import (
	"sync"
)

// ParserPoolRecycled may be used for pooling Parsers for structurally dissimilar JSON.
type ParserPoolRecycled struct {
	pool sync.Pool
	maxReuse int
}

// NewParserPoolRecycled enables JSON Parser pooling for semi-structured JSON
//
// MaxReuse prevents a parser from being returned to the pool after MaxReuse
// number of uses. This prevents parser reuse from causing unbounded memory
// growth for structurally dissimilar JSON. 1,000 is probably a good number.
func NewParserPoolRecycled(maxReuse int) *ParserPoolRecycled {
	return &ParserPoolRecycled{
		sync.Pool{New: func() interface{} { return new(ParserRecyclable) }},
		maxReuse,
	}
}

// Get returns a ParserRecyclable from ppr.
//
// The ParserRecyclable must be Put to ppr after use.
func (ppr *ParserPoolRecycled) Get() *ParserRecyclable {
	v := ppr.pool.Get()
	if v == nil {
		return &ParserRecyclable{}
	}
	return v.(*ParserRecyclable)
}

// Put returns pr to ppr.
//
// pr and objects recursively returned from pr cannot be used after pr
// is put into ppr.
func (ppr *ParserPoolRecycled) Put(pr *ParserRecyclable) {
	if pr.n > ppr.maxReuse {
		return
	}
	ppr.pool.Put(pr)
}

// ParserRecyclable adds a counter to a Parser for use with ParserPoolRecycled
type ParserRecyclable struct {
	Parser
	n int
}

// Parse is a wrapper for Parser.Parse that also counts the number of calls.
func (pr *ParserRecyclable) Parse(s string) (*Value, error) {
	pr.n++
	return pr.Parser.Parse(s)
}

// ParseBytes is a wrapper for Parser.ParseBytes that also counts the number of calls.
func (pr *ParserRecyclable) ParseBytes(b []byte) (*Value, error) {
	pr.n++
	return pr.Parser.ParseBytes(b)
}

// ScannerPoolRecycled may be used for pooling Scanners for structurally dissimilar JSON.
type ScannerPoolRecycled struct {
	pool sync.Pool
	maxReuse int
}

// NewScannerPoolRecycled enables JSON Scanner pooling for semi-structured JSON
//
// MaxReuse prevents a scanner from being returned to the pool after MaxReuse
// number of uses. This prevents scanner reuse from causing unbounded memory
// growth for structurally dissimilar JSON. 1,000 is probably a good number.
func NewScannerPoolRecycled(maxReuse int) *ScannerPoolRecycled {
	return &ScannerPoolRecycled{
		sync.Pool{New: func() interface{} { return new(ScannerRecyclable) }},
		maxReuse,
	}
}

// Get returns a ScannerRecyclable from spr.
//
// The ScannerRecyclable must be Put to spr after use.
func (spr *ScannerPoolRecycled) Get() *ScannerRecyclable {
	v := spr.pool.Get()
	if v == nil {
		return &ScannerRecyclable{}
	}
	return v.(*ScannerRecyclable)
}


// Put returns sr to spr.
//
// sr and objects recursively returned from sr cannot be used after sr
// is put into spr.
func (spr *ScannerPoolRecycled) Put(sr *ScannerRecyclable) {
	if sr.n > spr.maxReuse {
		return
	}
	spr.pool.Put(sr)
}

// ScannerRecyclable adds a counter to a Scanner for use with ScannerPoolRecycled
type ScannerRecyclable struct {
	Scanner
	n int
}

// Init is a wrapper for Scanner.Init that also counts the number of calls.
func (sr *ScannerRecyclable) Init(s string) {
	sr.n++
	sr.Scanner.Init(s)
}

// InitBytes is a wrapper for Scanner.InitBytes that also counts the number of calls.
func (sr *ScannerRecyclable) InitBytes(b []byte) {
	sr.n++
	sr.Scanner.InitBytes(b)
}
