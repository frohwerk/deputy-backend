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
		case v.ComponentId != "":
			sb.WriteString(v.ComponentId)
		default:
			sb.WriteString(v.ComponentId)
		}
		if i < limit {
			sb.WriteString(", ")
		}
	}
	return sb.String()
}
