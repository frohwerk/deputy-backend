package tasks

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"

	k8s "github.com/frohwerk/deputy-backend/internal/kubernetes"
)

type tracker map[string]map[string]bool

func (t tracker) AvailablePods(imageName string) uint {
	pods, ok := t[imageName]
	if !ok {
		fmt.Println("# of available pods for image", imageName, "is", 0)
		return 0
	}

	ct := uint(0)
	for _, available := range pods {
		if available {
			// fmt.Println("pod", podName, "running", imageName, "is available")
			ct++
		} else {
			// fmt.Println("pod", podName, "running", imageName, "is NOT available")
		}
	}
	return ct
}

func Apply(env *k8s.Environment, name, newimg string) <-chan interface{} {

	deployment, err := kube.AppsV1().Deployments("myproject").Get(name, meta.GetOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading deployment: %s\n", err)
		return
	}

	r := uint(*deployment.Spec.Replicas)
	fmt.Printf("Component %s uses %v replicas\n", deployment.Name, r)
	oldimg := deployment.Spec.Template.Spec.Containers[0].Image
	fmt.Printf("Updating image from %s to %s\n", oldimg, newimg)

	query := labels.Set(deployment.Spec.Selector.MatchLabels).String()
	podsWatch, err := kube.CoreV1().Pods("myproject").Watch(meta.ListOptions{LabelSelector: query})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting pods watch: %s\n", err)
		return
	}

	timeout := time.NewTimer(150 * time.Second)
	done := make(chan interface{}, 1)

	imagesPods := tracker{
		oldimg: {},
		newimg: {},
	}

	go func() {
	eventloop:
		for {
			select {
			case <-timeout.C:
				fmt.Println("Timeout waiting for updated pods")
				done <- nil
				break eventloop
			case evt := <-podsWatch.ResultChan():
				pod, ok := evt.Object.(*core.Pod)
				if !ok {
					fmt.Printf("unexpected object of type %T in %s-event", evt.Object, evt.Type)
					continue eventloop
				}

				img := pod.Spec.Containers[0].Image
				fmt.Printf("%v %s event for pod %s (running %s)\n", time.Now().UnixNano(), evt.Type, pod.Name, img)

				switch evt.Type {
				case watch.Added:
					fallthrough
				case watch.Modified:
					pods, ok := imagesPods[img]
					if !ok {
						fmt.Fprintf(os.Stderr, "unexpected image name in ADDED event: %s\n", img)
						continue
					}
					pods[pod.Name] = isPodAvailable(pod)
				}

				old := imagesPods.AvailablePods(oldimg)
				new := imagesPods.AvailablePods(newimg)
				fmt.Printf("%v # of available pods for image %s is %v/%v\n", time.Now().UnixNano(), oldimg, old, r)
				fmt.Printf("%v # of available pods for image %s is %v/%v\n", time.Now().UnixNano(), newimg, new, r)
				if old == 0 && new == r {
					done <- nil
					break eventloop
				}
			}
		}
	}()

	defer podsWatch.Stop()

	patch, err := json.Marshal(k8s.CreateImagePatch(name, newimg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating deployment patch: %s\n", err)
		return
	}

	fmt.Println("Patch looks like this:")
	fmt.Println(string(patch))

	if _, err := kube.AppsV1().Deployments("myproject").Patch(name, types.StrategicMergePatchType, patch); err != nil {
		fmt.Fprintf(os.Stderr, "error patching deployment: %s\n", err)
		return
	}

	<-done
}

func isPodAvailable(pod *core.Pod) bool {
	secs := pod.ObjectMeta.DeletionGracePeriodSeconds
	if secs != nil {
		return false // Pods scheduled for deletion are no longer used by load balancers
		// See https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination
	}
	for _, c := range pod.Status.Conditions {
		// fmt.Println("Pod", pod.Name, "Condition", c.Type, "is", c.Status)
		switch c.Type {
		case core.PodReady:
			return c.Status == core.ConditionTrue
		}
	}
	return false
}
