package tree

import (
	"fmt"

	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/frohwerk/deputy-backend/internal/util"
)

var Log logger.Logger = logger.Noop

var MaxDepth = 20

type Node struct {
	Value        string
	Depth        int
	Dependencies []*Node
}

type builder struct {
	cache  map[string]*Node
	lookup Lookup
}

type build struct {
	*builder
	recursion int
}

type NodeList []*Node

type Lookup func(id string) ([]string, error)

// Tree builder with caching support
func Builder(lookup Lookup) *builder {
	return &builder{cache: map[string]*Node{}, lookup: lookup}
}

func Create(rootId string, lookup Lookup) (*Node, error) {
	return Builder(lookup).CreateTree(rootId)
}

func (b *builder) CreateTree(rootId string) (*Node, error) {
	t := &build{builder: b, recursion: 0}
	root, err := t.createTree(rootId)
	if t.recursion != 0 {
		Log.Debug("recursion counter should be 0, shouldn't it? But it is %v", t.recursion)
	}
	return root, err
}

func (b *build) createTree(rootId string) (*Node, error) {
	b.recursion++
	defer func() { b.recursion-- }()

	if b.recursion > MaxDepth {
		return nil, fmt.Errorf("too many recursions, there might be a circular relationship or the tree is too deep. if you are sure it is not, then you can increase tree.MaxDepth")
	}

	if t, cached := b.cache[rootId]; cached {
		return t, nil
	}

	Log.Trace("Looking up dependencies for id: %s", rootId)
	deps, err := b.lookup(rootId)
	if err != nil {
		return nil, err
	}

	root := Node{Value: rootId, Depth: 0, Dependencies: []*Node{}}
	b.cache[rootId] = &root

	depth := 0
	for _, dep := range deps {
		node, err := b.createTree(dep)
		if err != nil {
			return nil, err
		}
		depth = util.MaxInt(depth, node.Depth)
		root.Dependencies = append(root.Dependencies, node)
	}

	if len(deps) > 0 {
		root.Depth = depth + 1
	}

	return &root, nil
}

type walker struct {
	seen map[string]*Node
}

func (n *Node) Children() []*Node {
	if n.Leaf() {
		return []*Node{}
	}

	w := &walker{seen: map[string]*Node{}}
	for _, n := range n.Dependencies {
		w.collectUnique(n)
	}

	result := []*Node{}
	for _, v := range w.seen {
		result = append(result, v)
	}

	return result
}

func (w *walker) collectUnique(n *Node) {
	switch {
	case n == nil:
		return
	case n.Leaf():
		w.seen[n.Value] = n
	default:
		for _, dependent := range n.Dependencies {
			w.collectUnique(dependent)
		}
	}
}

// Remove all nodes without subtrees and return these nodes. If the node is ignored, if you want to know if the node itself is a leaf use the Leaf() method
func (n *Node) Trim() []Node {
	t := &trimmer{processed: map[string]interface{}{}, removed: map[string]Node{}}
	t.trim(n)

	i, res := 0, make([]Node, len(t.removed))
	for _, n := range t.removed {
		res[i] = n
		i++
	}

	return res
}

type trimmer struct {
	processed map[string]interface{}
	removed   map[string]Node
}

func (t *trimmer) trim(n *Node) {
	if n == nil || n.Leaf() {
		return
	}

	Log.Trace("t.processed:")
	for k := range t.processed {
		Log.Trace("- %s", k)
	}
	Log.Trace("trim: %s", n.Value)
	for _, n := range n.Dependencies {
		Log.Trace("------> %s", n.Value)
	}
	for i := 0; i < len(n.Dependencies); {
		d := n.Dependencies[i]
		if _, ok := t.processed[d.Value]; ok {
			Log.Trace("%s already processed: %v", d.Value, ok)
			i++
		} else if d.Leaf() {
			t.removed[d.Value] = *slice(n, i)
		} else {
			t.trim(d)
			i++
		}
	}

	n.Depth--
	Log.Trace("t.processed <- %s", n.Value)
	t.processed[n.Value] = nil
}

func (n *Node) Leaf() bool {
	return n == nil || len(n.Dependencies) == 0
}

func (n *Node) String() string {
	return fmt.Sprintf("%s (%v nodes)", n.Value, n.Depth)
}

func slice(n *Node, i int) *Node {
	v := n.Dependencies[i]
	Log.Trace("slice: %s", v.Value)
	if i == len(n.Dependencies) {
		n.Dependencies = n.Dependencies[:i]
	} else {
		n.Dependencies = append(n.Dependencies[:i], n.Dependencies[i+1:]...)
	}
	return v
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
