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
	"k8s.io/client-go/kubernetes"

	k8s "github.com/frohwerk/deputy-backend/internal/kubernetes"
)

func Stuff(kube kubernetes.Interface, newimg string) {
	deployment, err := kube.AppsV1().Deployments("myproject").Get("node-hello-world", meta.GetOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading deployment: %s\n", err)
		return
	}

	r := int(*deployment.Spec.Replicas)
	fmt.Printf("Component %s uses %v replicas\n", deployment.Name, r)
	oldimg := deployment.Spec.Template.Spec.Containers[0].Image
	fmt.Printf("Updating image from %s to %s\n", oldimg, newimg)

	query := labels.Set(deployment.Spec.Selector.MatchLabels).String()
	podsWatch, err := kube.CoreV1().Pods("myproject").Watch(meta.ListOptions{LabelSelector: query})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error starting pods watch: %s\n", err)
		return
	}

	old := 0
	new := 0

	timeout := time.NewTimer(150 * time.Second)
	done := make(chan interface{}, 1)

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
				available := isPodAvailable(pod)
				fmt.Printf("%s %s Pod: %s Image: %s Phase: %s Available: %v\n", time.Now().Format(time.RFC3339), evt.Type, pod.Name, img, pod.Status.Phase, available)

				switch evt.Type {
				case watch.Added:
					switch img {
					case oldimg:
						old++
					case newimg:
						new++
					}
				case watch.Deleted:
					switch img {
					case oldimg:
						old--
					case newimg:
						new--
					}
				}
				fmt.Printf("%s running in %v/%v containers\n", oldimg, old, r)
				fmt.Printf("%s running in %v/%v containers\n", newimg, new, r)
				if new == r && old == 0 {
					done <- nil
				}
			}
		}
	}()

	defer podsWatch.Stop()

	patch, err := json.Marshal(k8s.CreateImagePatch("node-hello-world", newimg))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating deployment patch: %s\n", err)
		return
	}

	fmt.Println("Patch looks like this:")
	fmt.Println(string(patch))

	if _, err := kube.AppsV1().Deployments("myproject").Patch("node-hello-world", types.StrategicMergePatchType, patch); err != nil {
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
		fmt.Println("Pod", pod.Name, "Condition", c.Type, "is", c.Status)
		switch c.Type {
		case core.PodReady:
			return c.Status == core.ConditionTrue
		}
	}
	return false
}
