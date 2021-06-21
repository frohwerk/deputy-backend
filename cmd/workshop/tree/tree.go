package tree

import (
	"fmt"

	"github.com/frohwerk/deputy-backend/internal/logger"
)

var Log logger.Logger = logger.Noop

var MaxDepth = 20

type node struct {
	Value        string
	Dependencies []*node
}

type builder struct {
	depth  int
	cache  map[string]*node
	lookup Lookup
}

type Lookup func(id string) ([]string, error)

// Tree builder with caching support
func Builder(lookup Lookup) *builder {
	return &builder{cache: map[string]*node{}, lookup: lookup}
}

func Create(rootId string, lookup Lookup) (*node, error) {
	return Builder(lookup).CreateTree(rootId)
}

func (b *builder) CreateTree(rootId string) (*node, error) {
	b.depth++
	defer func() { b.depth-- }()

	if b.depth > MaxDepth {
		return nil, fmt.Errorf("too many recursions, there might be a circular relationship or the tree is too deep. if you are sure it is not, then you can increase tree.MaxDepth")
	}

	if t, cached := b.cache[rootId]; cached {
		return t, nil
	}

	Log.Debug("Looking up dependencies for id: %s", rootId)
	deps, err := b.lookup(rootId)
	if err != nil {
		return nil, err
	}

	root := node{Value: rootId, Dependencies: []*node{}}
	b.cache[rootId] = &root

	for _, dep := range deps {
		node, err := b.CreateTree(dep)
		if err != nil {
			return nil, err
		}
		root.Dependencies = append(root.Dependencies, node)
	}

	return &root, nil
}

// Search all nodes matching the match function
func (t *node) Search(match func(*node) bool) []*node {
	set := map[string]*node{}

	if t == nil {
		return nil
	}

	if match(t) {
		set[t.Value] = t
	}

	for _, dep := range t.Dependencies {
		matches := dep.Search(match)
		for _, item := range matches {
			set[item.Value] = item
		}
	}

	i := 0
	res := make([]*node, len(set))
	for _, v := range set {
		res[i] = v
		i++
	}

	return res
}

// Remove all nodes without subtrees and return these nodes. If the node is ignored, if you want to know if the node itself is a leaf use the Leaf() method
func (n *node) Trim() []node {
	t := &trimmer{removed: map[string]node{}}
	t.trim(n)

	i, res := 0, make([]node, len(t.removed))
	for _, n := range t.removed {
		res[i] = n
		i++
	}

	return res
}

type trimmer struct {
	removed map[string]node
}

func (t *trimmer) trim(n *node) {
	if n == nil {
		return
	}

	Log.Debug("trim: %s", n.Value)
	for i := 0; i < len(n.Dependencies); {
		d := n.Dependencies[i]
		if d.Leaf() {
			t.removed[d.Value] = *slice(n, i)
		} else {
			i++
			t.trim(d)
		}
	}
}

func (n *node) Leaf() bool {
	return n == nil || len(n.Dependencies) == 0
}

func slice(n *node, i int) *node {
	v := n.Dependencies[i]
	Log.Debug("slice: %s", v.Value)
	if i == len(n.Dependencies) {
		n.Dependencies = n.Dependencies[:i]
	} else {
		n.Dependencies = append(n.Dependencies[:i], n.Dependencies[i+1:]...)
	}
	return v
}
