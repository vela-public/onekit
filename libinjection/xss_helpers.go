package libinjection

import (
	"strings"
)

func isH5White(ch byte) bool {
	return ch == '\n' || ch == '\t' || ch == '\v' || ch == '\f' || ch == '\r' || ch == ' '
}

func isBlackTag(s string) bool {
	if len(s) < 3 {
		return false
	}

	sUpperWithoutNulls := strings.ToUpper(strings.ReplaceAll(s, "\x00", ""))
	for i := 0; i < len(blackTags); i++ {
		if sUpperWithoutNulls == blackTags[i] {
			return true
		}
	}

	switch sUpperWithoutNulls {
	// anything SVG or XSL(t) related
	case "SVT", "XSL":
		return true
	default:
		return false
	}
}

func isBlackAttr(s string) int {
	sUpperWithoutNulls := strings.ToUpper(strings.ReplaceAll(s, "\x00", ""))

	length := len(sUpperWithoutNulls)
	if length < 2 {
		return attributeTypeNone
	}
	if length >= 5 {
		if sUpperWithoutNulls == "XMLNS" || sUpperWithoutNulls == "XLINK" {
			// got xmlns or xlink tags
			return attributeTypeBlack
		}
		// JavaScript on.* event handlers
		if sUpperWithoutNulls[:2] == "ON" {
			eventName := sUpperWithoutNulls[2:]
			// got javascript on- attribute name
			for _, event := range blackEvents {
				if eventName == event.name {
					return event.attributeType
				}
			}
		}
	}

	for _, black := range blacks {
		if sUpperWithoutNulls == black.name {
			// got banner attribute name
			return black.attributeType
		}
	}
	return attributeTypeNone
}

func htmlDecodeByteAt(s string) (int, int) {
	length := len(s)
	val := 0

	if length == 0 {
		return byteEOF, 0
	}

	if s[0] != '&' || length < 2 {
		return int(s[0]), 1
	}

	if s[1] != '#' || len(s) < 3 {
		// normally this would be for named entities
		// but for this case we don't actually care
		return '&', 1
	}

	if s[2] == 'x' || s[2] == 'X' {
		if len(s) < 4 {
			return '&', 1
		}
		ch := int(s[3])
		ch = gsHexDecodeMap[ch]
		if ch == 256 {
			// degenerate case '&#[?]'
			return '&', 1
		}
		val = ch
		i := 4

		for i < length {
			ch = int(s[i])
			if ch == ';' {
				return val, i + 1
			}
			ch = gsHexDecodeMap[ch]
			if ch == 256 {
				return val, i
			}
			val = val*16 + ch
			if val > 0x1000FF {
				return '&', 1
			}
			i++
		}
		return val, i
	}
	i := 2
	ch := int(s[i])
	if ch < '0' || ch > '9' {
		return '&', 1
	}
	val = ch - '0'
	i++
	for i < length {
		ch = int(s[i])
		if ch == ';' {
			return val, i + 1
		}
		if ch < '0' || ch > '9' {
			return val, i
		}
		val = val*10 + (ch - '0')
		if val > 0x1000FF {
			return '&', 1
		}
		i++
	}
	return val, i
}

// Does an HTML encoded  binary string (const char*, length) start with
// a all uppercase c-string (null terminated), case insensitive!
//
// also ignore any embedded nulls in the HTML string!
func htmlEncodeStartsWith(a, b string) bool {
	var (
		first  = true
		bs     []byte
		pos    = 0
		length = len(b)
	)

	for length > 0 {
		cb, consumed := htmlDecodeByteAt(b[pos:])
		pos += consumed
		length -= consumed

		if first && cb <= 32 {
			// ignore all leading whitespace and control characters
			continue
		}
		first = false

		if cb == 0 || cb == 10 {
			// always ignore null characters in user input
			// always ignore vertical tab characters in user input
			continue
		}
		if cb >= 'a' && cb <= 'z' {
			cb -= 0x20
		}
		bs = append(bs, byte(cb))
	}

	return strings.Contains(string(bs), a)
}

func isBlackURL(s string) bool {
	urls := []string{
		"DATA",        // data url
		"VIEW-SOURCE", // view source url
		"VBSCRIPT",    // obsolete but interesting signal
		"JAVA",        // covers JAVA, JAVASCRIPT, + colon
	}

	//  HEY: this is a signed character.
	//  We are intentionally skipping high-bit characters too
	//  since they are not ASCII, and Opera sometimes uses UTF-8 whitespace.
	//
	//  Also in EUC-JP some of the high bytes are just ignored.
	str := strings.TrimLeftFunc(s, func(r rune) bool {
		return r <= 32 || r >= 127
	})

	for _, url := range urls {
		if htmlEncodeStartsWith(url, str) {
			return true
		}
	}
	return false
}
