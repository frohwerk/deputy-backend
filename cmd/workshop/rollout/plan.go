package rollout

import (
	"strings"
)

type Plan struct {
	queue PatchList
}

func (r *strategy) CreatePlan(source PatchList) (*Plan, error) {
	plan := &Plan{PatchList{}}
	Log.Debug("input: %s", source)
	for n := 0; len(source) > 0 && n < 10; n++ {
		Log.Debug("--- Slot #%v ----------------------------------------------------------------------", n)
		for i := 0; i < len(source); {
			c := source[i]
			deps, err := r.Lookup.Direct(c.ComponentId)
			if err != nil {
				return nil, err
			}
			Log.Debug("checking component <%s> with dependencies %s for slot #%v", c.ComponentId, deps, n)
			if plan.Satisfies(deps) {
				Log.Debug("dependencies for component <%s> are satisfied. moving to target slot #%v", c.ComponentId, n)
				plan.queue = append(plan.queue, c)
				source = append(source[:i], source[i+1:]...)
				Log.Trace("source: [%s] ||| plan: %s", source, plan)
			} else {
				i++
			}
		}
	}
	return plan, nil
}

func (plan *Plan) Satisfies(ids []string) bool {
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

func (plan *Plan) String() string {
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
