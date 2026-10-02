package fastjson

// Compact appends to dst the JSON-encoded src with insignificant space characters elided.
func Compact(dst, src []byte) ([]byte, error) {
	if err := ValidateBytes(src); err != nil {
		return dst, err
	}
	if cap(dst)-len(dst) < len(src) {
		dst2 := make([]byte, len(dst), len(dst)+len(src))
		copy(dst2, dst)
		dst = dst2
	}
	i := 0
	for i < len(src) {
		start := i
		for i < len(src) {
			c := src[i]
			if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
				if i > start {
					dst = append(dst, src[start:i]...)
				}
				i++
				for i < len(src) && (src[i] == ' ' || src[i] == '\t' || src[i] == '\r' || src[i] == '\n') {
					i++
				}
				start = i
				continue
			}
			if c == '"' {
				break
			}
			i++
		}
		if i > start {
			dst = append(dst, src[start:i]...)
		}
		if i >= len(src) {
			break
		}
		strStart := i
		i++
		for i < len(src) {
			c := src[i]
			if c == '\\' {
				i += 2
				continue
			}
			if c == '"' {
				i++
				break
			}
			i++
		}
		dst = append(dst, src[strStart:i]...)
	}
	return dst, nil
}
