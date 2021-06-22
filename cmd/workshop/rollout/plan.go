package rollout

type plan struct {
	slots []PatchList
}

func (r *strategy) CreatePlan(patches PatchList) (PatchList, error) {
	return nil, nil
}
