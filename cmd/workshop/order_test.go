package main

import (
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/stretchr/testify/assert"
)

func init() {
	Log = logger.Basic(logger.LEVEL_DEBUG)
}

type memoryStore map[string][]string

func (r *memoryStore) Direct(id string) ([]string, error) {
	if deps, ok := (*r)[id]; ok {
		return deps, nil
	}
	return []string{}, nil
}

func TestOrdering(t *testing.T) {
	var store = &memoryStore{
		"a": {"b"},
		"b": {"c", "d"},
		"d": {"f"},
	}
	cache := dependencies.Cache{Store: store}
	repo := dependencies.Lookup{Store: &cache}

	patches := PatchList{{ComponentId: "c"}, {ComponentId: "d"}, {ComponentId: "f"}, {ComponentId: "b"}}
	planner := &SortByDependencies{dependencies: repo}

	plan, err := planner.Plan(patches)
	if err != nil {
		t.Fatal("creating rollout plan failed:", err)
	}

	Log.Debug("final result: %s", plan)
	assert.Equal(t, "c", plan[0].ComponentId)
	assert.Equal(t, "f", plan[1].ComponentId)
	assert.Equal(t, "d", plan[2].ComponentId)
	assert.Equal(t, "b", plan[3].ComponentId)
}
