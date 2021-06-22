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
		plan := theplan{}
		source := createPatches("service-x", "frontend", "middleware", "service-y")
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware"},
			"middleware": {"service-x", "service-y"},
		})

		search := func(id string) []string {
			v, err := dependencies.Direct(id)
			if err != nil {
				t.Fatalf("error looking a dependencies for %s", id)
			}
			return v
		}

		magic(source, &plan, search)
	})

	t.Run("stuff #2", func(t *testing.T) {
		plan := &theplan{}
		source := createPatches("service-x", "service-y", "frontend", "middleware")
		dependencies := createLookup(memoryStore{
			"frontend":   {"middleware"},
			"middleware": {"service-x", "service-y"},
		})

		search := func(id string) []string {
			v, err := dependencies.Direct(id)
			if err != nil {
				t.Fatalf("error looking a dependencies for %s", id)
			}
			return v
		}

		magic(source, plan, search)
	})

}

func magic(source rollout.PatchList, plan *theplan, search func(id string) []string) {
	Log.Debug("input: %s", source)
	for n := 0; len(source) > 0 && n < 10; n++ {
		Log.Debug("--- Slot #%v ----------------------------------------------------------------------", n)
		plan.AddSlot()
		for i := 0; i < len(source); {
			c := source[i]
			deps := search(c.ComponentId)
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

	// sort dependencies within a slot
	for n := 1; n < len(plan.slots); n++ {
		slot := plan.slots[n]
		Log.Debug("checking slot #%v for internal dependencies: %s", n, slot)
		for i := 0; i < len(slot); i++ {
			j := i
			c := slot[i]
			deps := search(c.ComponentId)
			for _, dep := range deps {
				k := slot.Index(dep)
				Log.Trace("comparing %s at index %v <-> %s at index %v", slot[j].Name(), j, dep, k)
				if k > -1 && k < j {
					Log.Trace("swaping patches because of dependency")
					Log.Trace("before swap: [ %s ]", slot)
					Log.Trace("after swap:  [ %s ]", slot)
					slot[j], slot[k], j = slot[k], slot[j], k
				}
			}
		}
	}

	Log.Debug("source: [ %s ]", source)
	Log.Debug("plan:   [ %s ]", plan)
}

type theplan struct {
	slots []rollout.PatchList
}

func (plan *theplan) AddSlot() {
	plan.slots = append(plan.slots, rollout.PatchList{})
}

func (plan *theplan) AddPatch(p kubernetes.DeploymentPatch) {
	n := len(plan.slots) - 1
	plan.slots[n] = append(plan.slots[n], p)
}

func (plan *theplan) Satisfies(ids []string) bool {
	n := len(plan.slots) - 1
	Log.Trace("searching plan %s for dependencies %s", plan, ids)
	for i, slot := range plan.slots {
		if i == n {
			return len(ids) == 0
		}
		for _, comp := range slot {
			if comp.ComponentId == ids[0] {
				ids = ids[1:]
			}
			if len(ids) == 0 {
				return true
			}
		}
	}
	return len(ids) == 0
}

func (plan *theplan) String() string {
	sb := strings.Builder{}
	limit := len(plan.slots) - 1
	for i, slot := range plan.slots {
		sb.WriteString("[")
		sb.WriteString(slot.String())
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
