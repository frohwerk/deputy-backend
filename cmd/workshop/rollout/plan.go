package rollout

import (
	"fmt"
	"strings"
)

type builder struct {
	*strategy
	source PatchList
	queue  PatchList
}

func (s *strategy) CreatePlan(source PatchList) (PatchList, error) {
	plan := &builder{s, source, PatchList{}}
	Log.Debug("input: %s", source)
	for n := 0; len(source) > 0 && n < len(plan.source); n++ {
		Log.Debug("--- loop #%v ----------------------------------------------------------------------", n)
		Log.Debug("plan: [%s] | source: [%s]", plan, source)
		for i := 0; i < len(source); {
			c := source[i]
			deps, err := plan.dependencies(c.ComponentId)
			if err != nil {
				return nil, err
			}
			Log.Debug("checking component <%s> with dependencies %s for slot #%v", c.ComponentId, deps, n)
			if plan.queue.Contains(deps...) {
				Log.Debug("dependencies for component <%s> are satisfied. moving to target slot #%v", c.ComponentId, n)
				plan.queue = append(plan.queue, c)
				source = append(source[:i], source[i+1:]...)
			} else {
				i++
			}
		}
		if n == 0 && len(plan.queue) == 0 {
			return nil, fmt.Errorf("source [%s] contains a circular dependency", source)
		}
	}
	return plan.queue, nil
}

func (plan *builder) dependencies(id string) ([]string, error) {
	deps, err := plan.Lookup.Direct(id)
	if err != nil {
		return nil, err
	}
	Log.Trace("all dependencies for %s: %s", id, deps)
	for i := 0; i < len(deps); {
		Log.Trace("searching source [%s] for dependencies %s", plan.source, deps)
		if !plan.source.Contains(deps[i]) {
			Log.Trace("plan.source does not contain %s", deps[i])
			if last := len(deps) - 1; last > 0 {
				deps[i], deps[last] = deps[last], deps[i]
				deps = deps[:last]
			} else {
				return []string{}, nil
			}
		} else {
			i++
		}
	}

	Log.Trace("filtered dependencies for %s: %s", id, deps)
	return deps, err
}

func (plan *builder) contains(ids ...string) bool {
	return plan.queue.Contains(ids...)
}

func (plan *builder) String() string {
	sb := strings.Builder{}
	limit := len(plan.queue) - 1
	for i, patch := range plan.queue {
		sb.WriteString(patch.DisplayName())
		if i < limit {
			sb.WriteString(" -> ")
		}
	}
	return sb.String()
}
