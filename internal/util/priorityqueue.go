package util

import (
	"github.com/frohwerk/deputy-backend/internal/kubernetes"
)

type PriorityQueue struct {
	readable bool
	priority []int
	items    []*kubernetes.DeploymentPatch
}

func (queue *PriorityQueue) Len() int {
	return len(queue.items)
}

func (queue *PriorityQueue) Less(i, j int) bool {
	if queue.priority == nil {
		queue.priority = []int{}
	}
	return false
}

func (queue *PriorityQueue) Swap(i, j int) {

}

func (queue *PriorityQueue) Push() {
	if queue.items == nil {
		queue.items = []*kubernetes.DeploymentPatch{}
	}
}
