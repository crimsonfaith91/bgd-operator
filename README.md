# Blue-Green Deployment Operator

**This is not an actual Google product**

This repository implements a simple blue-green deployment operator using a CustomResourceDefinition (CRD).

## Running

```sh
# run the operator; kubeconfig is not required if operating in-cluster
$ go run *.go -kubeconfig=$HOME/.kube/config

# create a CustomResourceDefinition
$ kubectl create -f crd.yaml

# create a custom resource of type BGDeployment
$ kubectl create -f bgd.yaml

# check replicasets, pods, and service created through the custom resource
$ kubectl get all
```

When the `BGDeployment` custom resource is created, the operator will create a replicaset of 1 replica with `color=blue` label and a service with same color label.

## Image Change

The operator will create a new replicaset ONLY when there is a change of **.spec.image** field (image name with version) on custom resource template. Nothing happens for pod template change.

```sh
# edit the BGDeployment custom resource template
$ kubectl edit bgdeployment blue-green-deployment

# change .spec.image field to a different image name
# e.g., from "nginx:1.7.9" to "nginx:1.7.10"
```

If the image name with version is valid, the operator will create a new replicaset with different color label, alternating between blue and green colors. For example, if the old replicaset is labelled blue, a new replicaset labelled green will be created, and vice versa. When all pods of the new replicaset is ready and available, the operator will point the service to the new replicaset and tear down the old replicaset. 

If the image name with version is invalid, the operator will still create the new replicaset. However, the operator will soon find out that there is an `ErrImagePull` error and tear down the new replicaset. For example, if the old replicaset is labelled green, a new replicaset labelled blue with invalid image name will still be created. The operator will then wait for a certain timeout period for the new replicaset to be ready and available, and delete it when it fails to do so within the timeout period. The old replicaset and service stay intact like the new replicaset never exists.

## Cleanup

You can clean up the CRD with:

    $ kubectl delete crd bgdeployments.demo.google.com

CRD deletion cleans up the custom resource.

You can also clean up the custom resource with:

    $ kubectl delete bgdeployment blue-green-deployment

Custom resource deletion cleans up all resources (replicasets, pods, and service) created through it.

## References

* [sample-controller](https://github.com/kubernetes/kubernetes/tree/master/staging/src/k8s.io/sample-controller)
* [kube-crd](https://github.com/yaronha/kube-crd)