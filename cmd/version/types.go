package main

type DeploymentPatch struct {
	Spec DeploymentSpecPatch `json:"spec,omitempty"`
}

type DeploymentSpecPatch struct {
	Template PodTemplatePatch `json:"template,omitempty"`
}

type PodTemplatePatch struct {
	Spec PodSpecPatch `json:"spec,omitempty"`
}

type PodSpecPatch struct {
	Containers []ContainerPatch `json:"containers,omitempty" patchStrategy:"merge" patchMergeKey:"name"`
}

type ContainerPatch struct {
	Name  string `json:"name"`
	Image string `json:"image,omitempty"`
}
