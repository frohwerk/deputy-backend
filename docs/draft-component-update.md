Copy operation for one component:
old := deployments.Get(metadata.name)
replicas := old.Spec.Replicas
watch := pods.Watch(old.Spec.Selector.MatchLabels)
select {
case evt := watch.ResultChan():
    
}

Watch Pods
-> 

https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

A Pod must satisfy the Ready condition to be added to the load balancing pool
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-conditions

Pods that shut down slowly cannot continue to serve traffic as load balancers (like the service proxy) remove the Pod from the list of endpoints as soon as the termination grace period begins.
https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination
https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#objectmeta-v1-meta

deployment.Status.UnavailableReplicas > 0 && deployment.ObjectMetadata.Generation == deployment.Status.ObservedGeneration

deployment.Conditions.Progressing.LastUpdateTime might be the only timestamp usable as update timestamp

ImageID (with hash) is available once a Pod is modified to:
pod.Status.Conditions[type=ContainersReady].status=True