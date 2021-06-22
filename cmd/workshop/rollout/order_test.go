package rollout_test

import (
	"strings"
	"testing"

	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
	"github.com/frohwerk/deputy-backend/cmd/workshop/rollout"
	"github.com/frohwerk/deputy-backend/internal/kubernetes"
	"github.com/stretchr/testify/assert"
)

type memoryStore map[string][]string

func (r *memoryStore) Direct(id string) ([]string, error) {
	if deps, ok := (*r)[id]; ok {
		return deps, nil
	}
	return []string{}, nil
}

func init() {
	rollout.Log = Log
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

		plan, err := rollout.Strategy(dependencies).Plan(patches)
		if err != nil {
			t.Fatal("creating rollout plan failed:", err)
		}

		Log.Debug("final result: %s", plan)
		assert.Equal(t, "c", plan[0].ComponentId)
		assert.Equal(t, "f", plan[1].ComponentId)
		assert.Equal(t, "d", plan[2].ComponentId)
		assert.Equal(t, "b", plan[3].ComponentId)
	})

	t.Run("second test case", func(t *testing.T) {
		patches := createPatches("middleware", "service-x", "service-y", "frontend")
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware"},
			"middleware": {"service-x", "service-y"},
		})

		plan, err := rollout.Strategy(dependencies).Plan(patches)
		if assert.NoError(t, err, "creating rollout plan failed") {
			Log.Debug("final result: %s", plan)
			check := result{plan}
			check.Order(t, "middleware", "frontend")
			check.Order(t, "service-x", "middleware")
			check.Order(t, "service-y", "middleware")
		}
	})

	t.Run("third test case", func(t *testing.T) {
		patches := createPatches("service-x", "middleware", "service-y", "frontend")
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware"},
			"middleware": {"service-x", "service-y"},
		})

		plan, err := rollout.Strategy(dependencies).Plan(patches)
		if assert.NoError(t, err, "creating rollout plan failed") {
			Log.Debug("final result: %s", plan)
			check := result{plan}
			check.Order(t, "middleware", "frontend")
			check.Order(t, "service-x", "middleware")
			check.Order(t, "service-y", "middleware")
		}
	})

	t.Run("forth test case", func(t *testing.T) {
		patches := createPatches("service-x", "frontend", "middleware", "service-y")
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware"},
			"middleware": {"service-x", "service-y"},
		})

		plan, err := rollout.Strategy(dependencies).Plan(patches)
		if assert.NoError(t, err, "creating rollout plan failed") {
			Log.Debug("final result: %s", plan)
			check := result{plan}
			check.Order(t, "middleware", "frontend")
			check.Order(t, "service-x", "middleware")
			check.Order(t, "service-y", "middleware")
		}
	})

	t.Run("stuff #1", func(t *testing.T) {
		source := createPatches("service-x", "frontend", "middleware", "service-y")
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware"},
			"middleware": {"service-x", "service-y"},
		})
		m := &magician{dependencies}
		if plan, err := m.magic(source); assert.NoError(t, err) {
			Log.Debug("plan: [ %s ]", plan)
		}
	})

	t.Run("stuff #2", func(t *testing.T) {
		source := createPatches("service-x", "service-y", "frontend", "middleware")
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware"},
			"middleware": {"service-x", "service-y"},
		})
		m := &magician{dependencies}
		if plan, err := m.magic(source); assert.NoError(t, err) {
			Log.Debug("plan: [ %s ]", plan)
		}
	})

}

type magician struct {
	dependencies.Lookup
}

func (m *magician) magic(source rollout.PatchList) (*theplan, error) {
	plan := &theplan{rollout.PatchList{}}
	Log.Debug("input: %s", source)
	for n := 0; len(source) > 0 && n < 10; n++ {
		Log.Debug("--- Slot #%v ----------------------------------------------------------------------", n)
		for i := 0; i < len(source); {
			c := source[i]
			deps, err := m.Direct(c.ComponentId)
			if err != nil {
				return nil, err
			}
			Log.Debug("checking component <%s> with dependencies %s for slot #%v", c.ComponentId, deps, n)
			if plan.Satisfies(deps) {
				Log.Debug("dependencies for component <%s> are satisfied. moving to target slot #%v", c.ComponentId, n)
				plan.AddPatch(c)
				source = append(source[:i], source[i+1:]...)
				Log.Trace("source: [%s] ||| plan: %s", source, plan)
			} else {
				i++
			}
		}
	}

	// // sort dependencies within a slot
	// for n := 1; n < len(plan.things); n++ {
	// 	Log.Debug("checking slot #%v for internal dependencies: %s", n, plan.things)
	// 	for i := 0; i < len(plan.things); i++ {
	// 		j := i
	// 		c := slot[i]
	// 		deps, err := m.Direct(c.ComponentId)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		for _, dep := range deps {
	// 			k := slot.Index(dep)
	// 			Log.Trace("comparing %s at index %v <-> %s at index %v", slot[j].Name(), j, dep, k)
	// 			if k > -1 && k < j {
	// 				Log.Trace("swaping patches because of dependency")
	// 				Log.Trace("before swap: [ %s ]", slot)
	// 				Log.Trace("after swap:  [ %s ]", slot)
	// 				slot[j], slot[k], j = slot[k], slot[j], k
	// 			}
	// 		}
	// 	}
	// }

	Log.Debug("source: [ %s ]", source)

	return plan, nil
}

type theplan struct {
	queue rollout.PatchList
}

func (plan *theplan) AddPatch(p kubernetes.DeploymentPatch) {
	plan.queue = append(plan.queue, p)
}

func (plan *theplan) Satisfies(ids []string) bool {
	Log.Trace("searching plan %s for dependencies %s", plan, ids)
	for _, patch := range plan.queue {
		if len(ids) == 0 {
			return true
		}
		if patch.ComponentId == ids[0] {
			ids = ids[1:]
		}
	}
	return len(ids) == 0
}

func (plan *theplan) String() string {
	sb := strings.Builder{}
	limit := len(plan.queue) - 1
	for i, patch := range plan.queue {
		sb.WriteString("[")
		sb.WriteString(patch.Name())
		sb.WriteString("]")
		if i < limit {
			sb.WriteString(" -> ")
		}
	}
	return sb.String()
}

type result struct {
	rollout.PatchList
}

func (c result) Order(t *testing.T, a, b string) {
	x := c.PatchList.Index(a)
	y := c.PatchList.Index(b)
	assert.True(t, x > -1 && y > -1 && x < y, "%s should be before %s", a, b)
}
