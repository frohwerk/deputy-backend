package tasks

type deployment struct {
	Id            string
	ComponentName string
	PlatformName  string
	ImageRef      string
}

type deployments []deployment

func (d deployments) Len() int {
	return len(d)
}

func (d deployments) Less(i, j int) bool {
	return d[i].ComponentName < d[j].ComponentName
}

func (d deployments) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
