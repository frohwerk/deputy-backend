package tree_test

import (
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/tree"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
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
			assert.Len(t, root.Dependencies, 2)
			assert.Equal(t, "backend-for-frontend", root.Dependencies[0].Value)
			assert.Len(t, root.Dependencies[0].Dependencies, 1)
			assert.Equal(t, "backend", root.Dependencies[0].Dependencies[0].Value)
			assert.Equal(t, "backend", root.Dependencies[1].Value)
			assert.Len(t, root.Dependencies[1].Dependencies, 1)
			assert.Equal(t, "database", root.Dependencies[1].Dependencies[0].Value)
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
