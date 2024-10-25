package strkit

import (
	"unicode"
	"unicode/utf8"
)

type MatcherFSM struct {
	index   int
	matcher *Matcher
	current *trieNode
}

func (mf *MatcherFSM) Input(b rune) *Term {
	var term *Term

	char := unicode.ToLower(b)
	for mf.current.child[char] == nil && mf.current != mf.matcher.root {
		mf.current = mf.current.fail
	}

	if mf.current.child[char] != nil {
		mf.current = mf.current.child[char]
	}

	for p := mf.current; p != mf.matcher.root; p = p.fail {
		if p.count > 0 {
			for i := 0; i < p.count; i++ {
				term = &Term{Index: p.index, From: mf.index + utf8.RuneLen(char) - p.size, To: mf.index + utf8.RuneLen(char)}
				goto done
			}
		}
	}
done:
	mf.index += utf8.RuneLen(char)
	return term
}
