package rollout

import (
	"strings"

	"github.com/frohwerk/deputy-backend/internal/kubernetes"
)

type PatchList []kubernetes.DeploymentPatch

func (patches PatchList) Index(id string) int {
	for i, v := range patches {
		if v.ComponentId == id {
			return i
		}
	}
	return -1
}

func (patches PatchList) String() string {
	limit := len(patches) - 1
	sb := strings.Builder{}
	for i, v := range patches {
		switch {
		case v.ComponentName != "":
			sb.WriteString(v.ComponentName)
		default:
			sb.WriteString(v.ComponentId)
		}
		sb.WriteString(" [")
		sb.WriteString(v.Spec.Template.Spec.Containers[0].Image)
		sb.WriteString("]")
		if i < limit {
			sb.WriteString(" -> ")
		}
	}
	return sb.String()
}

func (patches PatchList) Contains(ids ...string) bool {
	Log.Trace("searching patches [%s] for %s", patches, ids)
	for _, patch := range patches {
		if len(ids) == 0 {
			return true
		}
		if patch.ComponentId == ids[0] {
			ids = ids[1:]
		}
	}
	return len(ids) == 0
}
