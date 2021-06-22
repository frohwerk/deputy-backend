package main

import (
	"strings"
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
	"github.com/frohwerk/deputy-backend/internal/kubernetes"
	"github.com/frohwerk/deputy-backend/internal/logger"
	"github.com/stretchr/testify/assert"
)

var Log = logger.Basic(logger.LEVEL_TRACE)

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
	rollout := &manager{dependencies: repo}

	plan, err := rollout.Plan(patches)
	if err != nil {
		t.Fatal("creating rollout plan failed:", err)
	}

	Log.Debug("final result: %s", plan)
	assert.Equal(t, "c", plan[0].ComponentId)
	assert.Equal(t, "f", plan[1].ComponentId)
	assert.Equal(t, "d", plan[2].ComponentId)
	assert.Equal(t, "b", plan[3].ComponentId)
}

func index(patches []kubernetes.DeploymentPatch, id string) int {
	for i, v := range patches {
		if v.ComponentId == id {
			return i
		}
	}
	return -1
}

type PatchList []kubernetes.DeploymentPatch

func (r *manager) Plan(patches PatchList) (PatchList, error) {
	plan := make(PatchList, len(patches))
	copy(plan, patches)
	for i, patch := range patches {
		Log.Trace("--- loop #%v ----------------------------------------------------------------------------------------", i)
		Log.Trace("original: %s", patches)
		Log.Trace("plan:     %s", plan)
		deps, err := r.dependencies.Transitive(patch.ComponentId)
		Log.Debug("dependencies of %s: %s", patch.ComponentId, deps)
		if err != nil {
			return nil, err
		}
		for _, dep := range deps {
			j := index(plan, dep)
			if j > -1 && i < j {
				Log.Debug(`%s depends on %s => swapping patches`, plan[i].ComponentId, plan[j].ComponentId)
				Log.Trace("before swap: %s", plan)
				plan[i], plan[j] = plan[j], plan[i]
				Log.Trace("after swap:  %s", plan)
			}
		}
	}
	return plan, nil
}

func (patches PatchList) String() string {
	limit := len(patches) - 1
	sb := strings.Builder{}
	for i, v := range patches {
		sb.WriteString(v.ComponentId)
		if i < limit {
			sb.WriteString(" -> ")
		}
	}
	return sb.String()
}

func Create(deps dependencies.Lookup, patches PatchList) (*manager, error) {
	return nil, nil
}

type manager struct {
	dependencies dependencies.Lookup
}

func (r *manager) Execute() {

}
