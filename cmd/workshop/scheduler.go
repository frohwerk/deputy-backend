package main

import (
	"strings"

	"github.com/frohwerk/deputy-backend/cmd/workshop/dependencies"
	"github.com/frohwerk/deputy-backend/internal/kubernetes"
	"github.com/frohwerk/deputy-backend/internal/logger"
)

var Log = logger.Default

type PatchList []kubernetes.DeploymentPatch

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

type SortByDependencies struct {
	dependencies dependencies.Lookup
}

func (r *SortByDependencies) Plan(patches PatchList) (PatchList, error) {
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
			j := plan.index(dep)
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

func (patches PatchList) index(id string) int {
	for i, v := range patches {
		if v.ComponentId == id {
			return i
		}
	}
	return -1
}
