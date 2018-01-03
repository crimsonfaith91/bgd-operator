# Blue-Green Deployment Operator

**This is not an actual Google product**

This repository implements a simple blue-green deployment operator using a CustomResourceDefinition (CRD). The operator maintains at most 2 replicasets (blue and green) at one time, alternating between the colors for new rollouts.

## Running

3 terminals are needed to run the operator locally (one for running local cluster, another for running the operator, and last one for interacting with the operator).

```sh
### first terminal ###

# install "bgd-operator" directory alongside with "kubernetes" directory

# navigate to "bgd-operator" directory
cd bgd-operator

# copy whole directory into main repo
cp -r . ../kubernetes/staging/src/k8s.io/bgd-operator

# navigate to "kubernetes" directory
cd ../kubernetes

# create a symlink in vendor package
ln -s ../../staging/src/k8s.io/bgd-operator vendor/k8s.io/bgd-operator

# start a local cluster
hack/local-up-cluster.sh

### second terminal ###

# navigate to "bgd-operator" directory
cd kubernetes/staging/src/k8s.io/bgd-operator

# run the operator; kubeconfig is not required if operating in-cluster
go run *.go -kubeconf=/var/run/kubernetes/admin.kubeconfig

### third terminal ###

# navigate to "bgd-operator" directory
cd kubernetes/staging/src/k8s.io/bgd-operator

# set up kubeconfig
export KUBECONFIG=/var/run/kubernetes/admin.kubeconfig

# create a CustomResourceDefinition
kubectl create -f crd.yaml

# create a custom resource of type BGDeployment
kubectl create -f bgd.yaml

# check replicasets, pods, and service created through the custom resource
kubectl get all
```

When the `BGDeployment` custom resource is created, the operator will create a replicaset of 1 replica with `color=blue` label and a service with same color label.

## Details

The operator will create a new replicaset ONLY when there is a change of **.spec.image** field (image name with version) on custom resource template. Nothing happens for pod template change.

```sh
### third terminal ###

# edit the BGDeployment custom resource template by changing ".spec.image" field
# e.g., from "nginx:1.7.9" to "nginx:1.7.10"
kubectl edit bgdeployment blue-green-deployment
```

Regardless a new rollout is successful or not, the operator will create a new replicaset. If the new rollout is successful (all pods of the new replicaset is ready and available within certain timeout period), the operator will point the service to the new replicaset and scale down the old replicaset to 0. Otherwise, it will scale down the new replicaset instead (the old replicaset and service stay intact). The zero-replica replicaset will be replaced during next successful rollout.

## Cleanup

You can clean up the CRD with:

    kubectl delete crd bgdeployments.demo.google.com

CRD deletion cleans up the custom resource.

You can also clean up the custom resource with:

    kubectl delete bgdeployment blue-green-deployment

Custom resource deletion cleans up replicasets and service created through it.

## Limitations

The operator does not support rollback. For example, if a user updates image name from `nginx:1.7.9` to `nginx:1.7.10` and back to `nginx:1.7.9` again, 2 rollouts will be performed resulting in 2 new replicasets being created.

The operator does not support some manual actions by the user, but this should not affect its main functionalities.
* When a replicaset is deleted manually, the operator will not respawn it and this will break the operator. This is because the operator only has a custom resource informer. 
* When an operator is turned off manually, all created resources will stay intact. The user has to manually delete all the resources before restarting the operator again, else there will be conflicts.

## References

* [sample-controller](https://github.com/kubernetes/kubernetes/tree/master/staging/src/k8s.io/sample-controller)
* [kube-crd](https://github.com/yaronha/kube-crd)
