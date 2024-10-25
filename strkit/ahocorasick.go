package strkit

import (
	"container/list"
	"github.com/vela-public/onekit/libkit"
	"unicode"
)

type trieNode struct {
	count int
	fail  *trieNode
	child map[rune]*trieNode
	index int
	size  int
}

func newTrieNode() *trieNode {
	return &trieNode{
		count: 0,
		fail:  nil,
		child: make(map[rune]*trieNode),
		index: -1,
	}
}

type Matcher struct {
	root *trieNode
	size int
	once bool
}

type Term struct {
	// indicates the index of the matching string in the original dictionary
	Index int

	//from
	From int
	// indicates the ending position index of the matched keyword in the input string s
	To int
}

func NewAcMatcher(once bool) *Matcher {
	return &Matcher{
		root: newTrieNode(),
		size: 0,
		once: once,
	}
}

func NewAcFromFile(path string, once bool) (*Matcher, error) {
	m := &Matcher{
		root: newTrieNode(),
		size: 0,
		once: once,
	}
	e := libkit.ReadlineFuncText(path, func(line string) (stop bool) {
		m.Insert(line)
		return false
	})
	if e != nil {
		return m, e
	}
	m.Build()
	return m, nil
}

func NewAc(dictionary []string, once bool) *Matcher {
	m := &Matcher{
		root: newTrieNode(),
		size: 0,
		once: once,
	}
	m.From(dictionary)
	return m
}

// initialize the ahocorasick
func (m *Matcher) From(dictionary []string) {
	for i := range dictionary {
		m.Insert(dictionary[i])
	}
	m.Build()
}

func (m *Matcher) Stream() *MatcherFSM {
	return &MatcherFSM{
		matcher: m,
		current: m.root,
		index:   0,
	}
}

// string match search
// return all strings matched as indexes into the original dictionary and their positions on matched string
func (m *Matcher) Match(s string) []*Term {
	fsm := m.Stream()
	var ret []*Term

	for _, char := range s {
		term := fsm.Input(unicode.ToLower(char))
		if term != nil {
			ret = append(ret, term)
			if m.once {
				return ret
			}
		}
	}

	return ret

	/*
		curNode := m.root
		var ret []*Term


		for index, char := range s {
			char = unicode.ToLower(char)
			for curNode.child[char] == nil && curNode != m.root {
				curNode = curNode.fail
			}

			if curNode.child[char] != nil {
				curNode = curNode.child[char]
			}
			for p := curNode; p != m.root; p = p.fail {
				if p.count > 0 {
					for i := 0; i < p.count; i++ {
						ret = append(ret, &Term{Index: p.index, From: index - p.size + 1, To: index + utf8.RuneLen(rune(char))})
						if m.once {
							return ret
						}
					}
				}
			}
		}

		return ret
	*/
}

func (m *Matcher) Build() {
	ll := list.New()
	ll.PushBack(m.root)
	for ll.Len() > 0 {
		temp := ll.Remove(ll.Front()).(*trieNode)

		for i, v := range temp.child {
			if temp == m.root {
				v.fail = m.root
			} else {
				p := temp.fail
				for p != nil {
					if childNode, ok := p.child[i]; ok {
						v.fail = childNode
						break
					}
					p = p.fail
				}
				if p == nil {
					v.fail = m.root
				}
			}
			ll.PushBack(v)
		}
	}
}

func (m *Matcher) Insert(s string) {
	if len(s) == 0 {
		return
	}

	curNode := m.root
	for _, char := range s {
		char = unicode.ToLower(char)
		if curNode.child[char] == nil {
			curNode.child[char] = newTrieNode()
		}
		curNode = curNode.child[char]
	}
	curNode.count++
	curNode.index = m.size
	curNode.size = len(s)
	m.size++

}
