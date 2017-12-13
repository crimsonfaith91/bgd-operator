# Blue-Green Deployment Operator

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

## Cleanup

You can clean up the CRD with:

    $ kubectl delete crd bgdeployments.demo.crimsonfaith91.com

CRD deletion cleans up the custom resource.

You can also clean up the custom resource with:

    $ kubectl delete bgd blue-green-deployment

Custom resource deletion cleans up all resources (replicasets, pods, and service) created through it.

## References

* [sample-controller](https://github.com/kubernetes/kubernetes/tree/master/staging/src/k8s.io/sample-controller)
* [kube-crd](https://github.com/yaronha/kube-crd)