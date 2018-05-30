package fastjson

var handyPool ParserPool

// GetString returns string value for the field identified by keys path
// in JSON data.
//
// Array indexes may be represented as decimal numbers in keys.
//
// An empty string is returned on error. Use Parser for proper error handling.
func GetString(data []byte, keys ...string) string {
	p := handyPool.Get()
	v, err := p.ParseBytes(data)
	if err != nil {
		handyPool.Put(p)
		return ""
	}
	sb := v.GetStringBytes(keys...)
	str := string(sb)
	handyPool.Put(p)
	return str
}

// GetBytes returns string value for the field identified by keys path
// in JSON data.
//
// Array indexes may be represented as decimal numbers in keys.
//
// nil is returned on error. Use Parser for proper error handling.
func GetBytes(data []byte, keys ...string) []byte {
	p := handyPool.Get()
	v, err := p.ParseBytes(data)
	if err != nil {
		handyPool.Put(p)
		return nil
	}
	sb := v.GetStringBytes(keys...)

	// Make a copy of sb, since sb belongs to p.
	var b []byte
	if sb != nil {
		b = append(b, sb...)
	}

	handyPool.Put(p)
	return b
}

// GetInt returns int value for the field identified by keys path
// in JSON data.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned on error. Use Parser for proper error handling.
func GetInt(data []byte, keys ...string) int {
	p := handyPool.Get()
	v, err := p.ParseBytes(data)
	if err != nil {
		handyPool.Put(p)
		return 0
	}
	n := v.GetInt(keys...)
	handyPool.Put(p)
	return n
}

// GetFloat64 returns float64 value for the field identified by keys path
// in JSON data.
//
// Array indexes may be represented as decimal numbers in keys.
//
// 0 is returned on error. Use Parser for proper error handling.
func GetFloat64(data []byte, keys ...string) float64 {
	p := handyPool.Get()
	v, err := p.ParseBytes(data)
	if err != nil {
		handyPool.Put(p)
		return 0
	}
	f := v.GetFloat64(keys...)
	handyPool.Put(p)
	return f
}

// GetBool returns boolean value for the field identified by keys path
// in JSON data.
//
// Array indexes may be represented as decimal numbers in keys.
//
// False is returned on error. Use Parser for proper error handling.
func GetBool(data []byte, keys ...string) bool {
	p := handyPool.Get()
	v, err := p.ParseBytes(data)
	if err != nil {
		handyPool.Put(p)
		return false
	}
	b := v.GetBool(keys...)
	handyPool.Put(p)
	return b
}

// Exists returns true if the field identified by keys path exists in JSON data.
//
// Array indexes may be represented as decimal numbers in keys.
func Exists(data []byte, keys ...string) bool {
	p := handyPool.Get()
	v, err := p.ParseBytes(data)
	if err != nil {
		handyPool.Put(p)
		return false
	}
	ok := v.Exists(keys...)
	handyPool.Put(p)
	return ok
}
