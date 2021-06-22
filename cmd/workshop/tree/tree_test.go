package tree_test

import (
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/tree"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	tree.Log = logger.Basic(logger.LEVEL_DEBUG)
}

func TestTree(t *testing.T) {
	dependencies := map[string][]string{
		"backend-for-frontend": {"backend"},
		"frontend":             {"backend-for-frontend", "backend"},
		"backend":              {"database"},
	}

	lookup := func(s string) ([]string, error) {
		if v, ok := dependencies[s]; ok {
			return v, nil
		}
		return []string{}, nil
	}

	t.Run("basic usage", func(t *testing.T) {
		root, err := tree.Create("frontend", lookup)
		if assert.NoError(t, err) {
			assert.Equal(t, "frontend", root.Value)
			assert.Equal(t, 3, root.Depth)
			assert.Len(t, root.Dependencies, 2)

			assert.Equal(t, "backend-for-frontend", root.Dependencies[0].Value)
			assert.Equal(t, 2, root.Dependencies[0].Depth)
			assert.Len(t, root.Dependencies[0].Dependencies, 1)

			assert.Equal(t, "backend", root.Dependencies[0].Dependencies[0].Value)
			assert.Equal(t, 1, root.Dependencies[0].Dependencies[0].Depth)

			assert.Equal(t, "backend", root.Dependencies[1].Value)
			assert.Equal(t, 1, root.Dependencies[1].Depth)
			assert.Len(t, root.Dependencies[1].Dependencies, 1)

			assert.Equal(t, "database", root.Dependencies[1].Dependencies[0].Value)
			assert.Equal(t, 0, root.Dependencies[1].Dependencies[0].Depth)
			assert.Len(t, root.Dependencies[1].Dependencies[0].Dependencies, 0)
		}
	})

	t.Run("infinite recursion protection", func(t *testing.T) {
		tree.MaxDepth = 1
		_, err := tree.Create("frontend", lookup)
		t.Log("expected error:", err)
		assert.Error(t, err, "expected an error, because the tree should be deeper than tree.MaxDepth = 1")
	})
}

func TestTrimming(t *testing.T) {
	tree.MaxDepth = 20

	dependencies := map[string][]string{
		"x": {"y", "z"},
		"y": {"a"},
		"z": {"a"},
		"a": {"b"},
	}

	lookup := func(s string) ([]string, error) {
		if v, ok := dependencies[s]; ok {
			return v, nil
		}
		return []string{}, nil
	}

	t.Run("trimming", func(t *testing.T) {
		x, err := tree.Create("x", lookup)
		if assert.NoError(t, err) {
			assert.Equal(t, "x", x.Value)
			assert.Equal(t, 3, x.Depth)
			assert.Len(t, x.Dependencies, 2)

			y := x.Dependencies[0]
			assert.Equal(t, "y", y.Value)
			assert.Equal(t, 2, y.Depth)
			assert.Len(t, y.Dependencies, 1)

			z := x.Dependencies[1]
			assert.Equal(t, "z", z.Value)
			assert.Equal(t, 2, z.Depth)
			assert.Len(t, z.Dependencies, 1)

			a := y.Dependencies[0]
			assert.Equal(t, "a", a.Value)
			assert.Equal(t, 1, a.Depth)
			assert.Len(t, a.Dependencies, 1)
			assert.Same(t, a, z.Dependencies[0])

			b := a.Dependencies[0]
			assert.Equal(t, "b", b.Value)
			assert.Equal(t, 0, b.Depth)
			assert.Len(t, b.Dependencies, 0)

			nodes := x.Trim()
			if assert.Len(t, nodes, 1) {
				assert.Equal(t, "b", nodes[0].Value)
				assert.Equal(t, 0, nodes[0].Depth)
				assert.Len(t, nodes[0].Dependencies, 0)
				assert.Equal(t, 2, x.Depth)
				assert.Equal(t, 1, y.Depth)
				assert.Equal(t, 1, z.Depth)
				assert.Equal(t, 0, a.Depth)
				assert.Equal(t, 0, b.Depth)
			}
		}
	})
}

// t
//  x
//   a
//    b
//  y
//   a
//    b
