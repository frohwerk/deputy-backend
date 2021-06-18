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
