package rollout

import (
	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
)

func Strategy(lookup dependencies.Lookup) *strategy {
	return &strategy{Lookup: lookup}
}

type strategy struct {
	Lookup dependencies.Lookup
}

func (r *strategy) Plan(patches PatchList) (PatchList, error) {
	if len(patches) == 0 {
		return PatchList{}, nil
	}
	plan := make(PatchList, len(patches))
	copy(plan, patches)
	Log.Trace("original: %s", patches)
	for i, patch := range patches {
		Log.Trace("--- loop #%v: %s --------------------------------------------------------------------------------------", i, patch.Name())
		Log.Trace("plan:     %s", plan)
		deps, err := r.Lookup.Direct(patch.ComponentId)
		Log.Debug("dependencies of %s: %s", patch.Name(), deps)
		if err != nil {
			return nil, err
		}
		j := plan.Index(patch.ComponentId) // current component will move around, but the iteration index may not...
		for _, dep := range deps {
			k := plan.Index(dep)
			if k == -1 {
				continue
			}
			Log.Trace("comparing %s at index %v <-> %s at index %v", plan[j].Name(), j, plan[k].Name(), k)
			if k > -1 && j < k {
				Log.Debug(`%v: %s depends on %s => swapping patches`, j, plan[j].Name(), plan[k].Name())
				Log.Trace("before swap: %s", plan)
				plan[j], plan[k] = plan[k], plan[j]
				Log.Trace("after swap:  %s", plan)
				j = k // update index for current component...
			}
		}
	}
	return plan, nil
}
