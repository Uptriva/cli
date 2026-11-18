package cmd

import "strings"

func sanitizeBody(b []byte) string {
	return strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return r
		}
		if r < 0x20 || r == 0x7f || r == 0x9b {
			return -1
		}
		return r
	}, string(b))
}

func safe(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\t' {
			return r
		}
		if r < 0x20 || r == 0x7f || r == 0x9b || r == 0x2028 || r == 0x2029 {
			return -1
		}
		return r
	}, s)
}
