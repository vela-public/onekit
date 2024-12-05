package treekit

import (
	"github.com/vela-public/onekit/cast"
	"strconv"
	"strings"
)

type TreeNode[T any] struct {
	Root     bool
	End      bool
	EndN     int
	Segment  string
	DataLog  T
	Children []*TreeNode[T]
}

func NewTree[T any](start string) *TreeNode[T] {
	return &TreeNode[T]{
		Root:    true,
		Segment: start,
	}
}

func (t *TreeNode[T]) Upsert(segment string, end bool) *TreeNode[T] {
	var current *TreeNode[T]
	for _, child := range t.Children {
		if child.Segment == segment {
			child.End = end
			current = child
			return current
		}
	}

	current = &TreeNode[T]{
		Segment: segment,
		End:     end,
		Root:    false,
	}

	t.Children = append(t.Children, current)
	return current
}

func (t *TreeNode[T]) Uri(uri string) *T { // /v1/api/info/app => v1 api info app

	segments := strings.Split(uri, "/")
	nf := len(segments) - 1

trim:
	if nf > 0 && segments[nf] == "" {
		segments = segments[:nf]
		nf = len(segments) - 1
		goto trim
	}

	current := t
	for i, seg := range segments {
		if seg == "" {
			if i == 0 {
				current = current.Upsert("/", i == nf)
			}
		} else {
			current = current.Upsert(seg, i == nf)
		}

		if i == nf-1 {
			current.EndN++
		}
	}
	return &current.DataLog
}

type TreeNodeFSM struct {
	base   int
	name   string
	parent string
}

func (fsm *TreeNodeFSM) incr() string {
	fsm.base++
	return strconv.Itoa(fsm.base)
}

func (fsm *TreeNodeFSM) ID() string {
	return strconv.Itoa(fsm.base)
}

func (t *TreeNode[T]) Label() string {
	return strconv.Quote(t.Segment)
}

func (t *TreeNode[T]) xLabel() string {
	return strconv.Quote(cast.ToString(t.DataLog))
}

func (t *TreeNode[T]) Attrs() map[string]string {
	return map[string]string{
		"label":         t.Label(),
		"tooltip":       t.xLabel(),
		"labelangle":    "45",
		"labeldistance": "2",
	}
}

/*

func (t *TreeNode[T]) Graphviz(w io.Writer) {

	fsm := &TreeNodeFSM{
		base:   0,
		name:   "sitemap",
		parent: "root",
		graph:  gographviz.NewGraph(),
	}
	_ = fsm.graph.SetName(fsm.name)
	_ = fsm.graph.SetDir(true)
	_ = fsm.graph.AddNode(fsm.name, fsm.ID(), t.Attrs())
	t.graphviz(fsm)
	_, _ = w.Write(cast.S2B(fsm.graph.String()))
}

func (t *TreeNode[T]) graphviz(fsm *TreeNodeFSM) {
	parent := fsm.ID()
	for _, child := range t.Children {
		id := fsm.incr()
		_ = fsm.graph.AddNode("sitemap", id, child.Attrs())
		_ = fsm.graph.AddEdge(parent, id, true, nil)
		child.graphviz(fsm)
	}
}
*/
