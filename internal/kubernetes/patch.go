package kubernetes

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type DeploymentPatch struct {
	Component string              `json:"-"`
	Platform  string              `json:"-"`
	Spec      DeploymentPatchSpec `json:"spec,omitempty"`
}

type DeploymentPatchSpec struct {
	Template PodTemplatePatch `json:"template,omitempty"`
}

type PodTemplatePatch struct {
	Spec corev1.PodSpec `json:"spec,omitempty"`
}

func Apply(target *platform, patch *DeploymentPatch) error {
	deployment, err := target.Deployments().Get(patch.Component, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("error reading deployment on target platform: %v", err)
	}

	replicas := uint(*deployment.Spec.Replicas)
	fmt.Printf("Component %s uses %v replicas\n", deployment.Name, replicas)
	currentImage := deployment.Spec.Template.Spec.Containers[0].Image
	fmt.Printf("Updating image from %s to %s\n", currentImage, patch.Spec.Template.Spec.Containers[0].Image)

	query := labels.Set(deployment.Spec.Selector.MatchLabels).String()
	watch, err := target.Pods().Watch(metav1.ListOptions{LabelSelector: query})
	if err != nil {
		return fmt.Errorf("error starting pods watch: %v", err)
	}

	// TODO Get things done (track the number of old and new images available, block until done) - timeout?
	return nil
}

func CreateImagePatch(container, imageRef string) DeploymentPatch {
	return DeploymentPatch{
		Spec: DeploymentPatchSpec{
			Template: PodTemplatePatch{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: container, Image: imageRef}},
				},
			},
		},
	}
}
