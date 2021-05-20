package tasks

type deployment struct {
	Name     string
	ImageRef string
}

type deployments []deployment

func (d deployments) Len() int {
	return len(d)
}

func (d deployments) Less(i, j int) bool {
	return d[i].Name < d[j].Name
}

func (d deployments) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
