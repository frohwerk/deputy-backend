package kubernetes

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
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

type images map[string]pods

type pods map[string]bool

func (p pods) Available() uint {
	ct := uint(0)
	for _, available := range p {
		if available {
			ct++
		}
	}
	return ct
}

func Apply(target *platform, patch *DeploymentPatch) (<-chan interface{}, error) {
	deployment, err := target.Deployments().Get(patch.Component, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error reading deployment on target platform: %v", err)
	}

	replicas := uint(*deployment.Spec.Replicas)
	fmt.Printf("Component %s uses %v replicas\n", deployment.Name, replicas)
	oldImage := deployment.Spec.Template.Spec.Containers[0].Image
	newImage := patch.Spec.Template.Spec.Containers[0].Image
	fmt.Printf("Updating image from %s to %s\n", oldImage, newImage)

	query := labels.Set(deployment.Spec.Selector.MatchLabels).String()
	podsw, err := target.Pods().Watch(metav1.ListOptions{LabelSelector: query})
	if err != nil {
		return nil, fmt.Errorf("error starting pods watch: %v", err)
	}

	defer podsw.Stop()

	complete := make(chan interface{}, 1)
	podsWith := images{
		oldImage: pods{},
		newImage: pods{},
	}

	go func() {
	eventloop:
		for evt := range podsw.ResultChan() {
			pod, ok := evt.Object.(*corev1.Pod)
			if !ok {
				fmt.Printf("unexpected object of type %T in %s-event", evt.Object, evt.Type)
				continue eventloop
			}

			image := pod.Spec.Containers[0].Image
			// fmt.Printf("%v %s event for pod %s (running %s)\n", time.Now().UnixNano(), evt.Type, pod.Name, img)
			if evt.Type == watch.Added || evt.Type == watch.Modified {
				pods, ok := podsWith[image]
				if !ok {
					fmt.Printf("unexpected image name in %s event: %s", evt.Type, image)
					continue eventloop
				}
				pods[pod.Name] = isPodAvailable(pod)
			}
			if podsWith[oldImage].Available() == 0 && podsWith[newImage].Available() == replicas {
				complete <- nil
			}
		}
	}()
	// TODO Get things done (track the number of old and new images available, block until done) - timeout?
	// TODO: wait for ${replicas} updated pods to complete startup and ${replicas} old pods to be scheduled for deletion
	patchData, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("error marshalling patch data: %v", err)
	}

	target.Deployments().Patch(patch.Component, types.StrategicMergePatchType, patchData)

	return complete, nil
}

func isPodAvailable(pod *corev1.Pod) bool {
	secs := pod.ObjectMeta.DeletionGracePeriodSeconds
	if secs != nil {
		return false // Pods scheduled for deletion are no longer used by load balancers
		// See https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination
	}
	for _, c := range pod.Status.Conditions {
		switch c.Type {
		case corev1.PodReady:
			return c.Status == corev1.ConditionTrue
		}
	}
	return false
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
