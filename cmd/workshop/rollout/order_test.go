package rollout_test

import (
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
	"github.com/frohwerk/deputy-backend/cmd/workshop/rollout"
	"github.com/frohwerk/deputy-backend/internal/kubernetes"
	"github.com/stretchr/testify/assert"
)

func init() {
	rollout.Log = Log
}

type memoryStore map[string][]string

func (r *memoryStore) Direct(id string) ([]string, error) {
	if deps, ok := (*r)[id]; ok {
		return deps, nil
	}
	return []string{}, nil
}

func TestOrdering(t *testing.T) {
	createLookup := func(v memoryStore) dependencies.Lookup {
		cache := dependencies.Cache(&v)
		return dependencies.Lookup{Store: cache}
	}

	createPatches := func(v ...string) rollout.PatchList {
		l := make(rollout.PatchList, len(v))
		for i := 0; i < len(v); i++ {
			l[i] = kubernetes.DeploymentPatch{ComponentId: v[i]}
		}
		return l
	}

	t.Run("basic use case", func(t *testing.T) {
		patches := createPatches("c", "d", "f", "b")
		dependencies := createLookup(memoryStore{
			"a": {"b"},
			"b": {"c", "d"},
			"d": {"f"},
		})

		plan, err := rollout.Strategy(dependencies).CreatePlan(patches)
		if assert.NoError(t, err, "creating rollout plan failed") {
			Log.Debug("plan: %s", plan)
			assert.Equal(t, "c", plan[0].ComponentId)
			assert.Equal(t, "f", plan[1].ComponentId)
			assert.Equal(t, "d", plan[2].ComponentId)
			assert.Equal(t, "b", plan[3].ComponentId)
		}
	})

	standardTest := func(t *testing.T, cases []rollout.PatchList, dependencies dependencies.Lookup) {
		c := make([]rollout.PatchList, len(cases))
		copy(c, cases)
		for _, patches := range c {
			t.Run(patches.String(), func(t *testing.T) {
				plan, err := rollout.Strategy(dependencies).CreatePlan(patches)
				if assert.NoError(t, err, "creating rollout plan failed") {
					check := result{plan}
					Log.Debug("plan: %s", plan)
					check.Order(t, "middleware", "frontend")
					check.Order(t, "service-x", "middleware")
					check.Order(t, "service-y", "middleware")
				}
			})
		}
	}

	cases := func() []rollout.PatchList {
		return []rollout.PatchList{
			createPatches("middleware", "service-x", "service-y", "frontend"),
			createPatches("service-x", "middleware", "service-y", "frontend"),
			createPatches("service-x", "frontend", "middleware", "service-y"),
			createPatches("service-x", "service-y", "frontend", "middleware"),
		}
	}

	t.Run("standard cases", func(t *testing.T) {
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware"},
			"middleware": {"service-x", "service-y"},
		})
		standardTest(t, cases(), dependencies)
	})

	t.Run("unused dependencies", func(t *testing.T) {
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware", "service-b"},
			"middleware": {"service-x", "service-y", "service-z"},
			"service-a":  {"service-b"},
		})
		standardTest(t, cases(), dependencies)
	})

	t.Run("circular dependency #1", func(t *testing.T) {
		patches := createPatches("a", "b", "c")
		dependencies := createLookup(memoryStore{
			"a": {"b"},
			"b": {"c"},
			"c": {"a"},
		})
		_, err := rollout.Strategy(dependencies).CreatePlan(patches)
		assert.Error(t, err, "should detect circular dependency a -> b -> c -> a")
	})

	t.Run("circular dependency #2", func(t *testing.T) {
		patches := createPatches("a", "b", "c", "d")
		dependencies := createLookup(memoryStore{
			"a": {"b"},
			"b": {"c", "d"},
			"c": {"d"},
			"d": {"c"},
		})
		_, err := rollout.Strategy(dependencies).CreatePlan(patches)
		assert.Error(t, err, "should detect circular dependency c <-> d")
	})

}

type result struct {
	plan rollout.PatchList
}

func (c result) Order(t *testing.T, a, b string) {
	x := c.plan.Index(a)
	y := c.plan.Index(b)
	assert.True(t, x > -1 && y > -1 && x < y, "%s should be before %s", a, b)
}
