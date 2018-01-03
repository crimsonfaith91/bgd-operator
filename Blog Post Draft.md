# Custom Controller, Custom Resource Definition and Operator

One reason Kubernetes succeeds in becoming new norm for deploying both on-premise and cloud applications is its API extensibility. With some background knowledge on how controllers work, it is straight-forward to write custom controllers that enjoy full-fledged capabilities like built-in controllers. Besides, Kubernetes also grants users ability to define custom resources based on application needs that the custom controllers can handle with.

A `custom resource` is a user-customized resource that allows users to store and retrieve structured data. The Kubernetes API is able to handle storage of custom resources in the same manner as built-in resources. The users can also use `kubectl` to perform CRUD operations on the custom resources.

Kubernetes provides ability of defining the custom resources via `Custom Resource Definition (CRD)`. A CRD is a custom resource template. It specifies name and schema of a custom resource. You can think that CRD is like a class definition in Java / C++, and a custom resource is like an object created based on the class definition.

A `custom controller` is a user provided controller that works with both built-in and custom resources. An `operator` is created when combining the custom controller with custom resources. The operator utilizes domain and application-specific knowledge to manage applications, greatly simplifying the task to apply human operational knowledge for cluster maintenance and monitoring purposes.

# Application: Blue-Green Deployment

<p align="center">
	<img src="http://cdn.ttgtmedia.com/rms/onlineImages/bluegreen_deployment_desktop.jpg">
	<br>
	<i>source: http://searchitoperations.techtarget.com/definition/blue-green-deployment</i>
</p>

Writing a Kubernetes operator can be tricky for first-time learners. To demonstrate how to write the operator, remaining blog post will go over a real-world use case named `blue-green deployment`. A blue-green deployment eliminates downtime and reduces risk by running at most two deployment versions (Blue and Green). To guarantee availability, one of the deployment versions must be alive at all times to serve production traffic.

Initially, we only have one release (let's mark it as `Blue` version). When rolling out a new release, we deploy and test a new (`Green`) version without tearing down the `Blue` version i.e., the `Green` version co-exists with the `Blue` version. After ensuring the `Green` version is working correctly, we redirect production traffic to it gradually. If something unexpected happens, we can then redirect traffic back to the `Blue` version immediately. When rolling out next new (3rd) release, we tear down the older (`Blue`) version before deploying the newest version. This time, the `Green` version becomes the `Blue` version and the newest version becomes the `Green` version. Same steps apply to future releases. A real world application of blue-green deployment is to carry out `A/B testing` for a web application.

Kubernetes does not support blue-green deployment natively, but its API allows creating a blue-green deployment operator.

# Writing the Operator

A basic blue-green deployment operator should comprise of a custom controller that handles two `ReplicaSets` (`blueRS` representing old deployment version and `greenRS` representing new deployment version) and a `service` redirecting traffic to newest version.

## Custom Controller

The custom controller should create:
1. a CRD based on given CRD YAML file
2. a `BGDeployment` custom resource based on given custom resource YAML file
3. a service pointing to first (`Blue`) deployment version
4. a client performing CRUD operations to the `BGDeployment` custom resource
5. an informer listening to `create`, `update`, and `delete` events of the `BGDeployment` custom resource

The client needs to define `Create`, `Get`, `Update`, `Delete`, `List` and `ListWatch` functions targeting the `BGDeployment` custom resource. Starting v1.8, you can utilize `client-gen` library to create a native typed client for the custom resources. Refer this [article](https://blog.openshift.com/kubernetes-deep-dive-code-generation-customresources/) for more information.

Initially, the informer will create `Blue` ReplicaSet `blueRS` and service via `AddFunc` event handler function. The service will be set to point to `blueRS`. `blueRS` is treated as available when all its pods are available (i.e., the pods have been ready for at least `MinReadySeconds`). The operator will then wait for any change to `.spec.image` field of the `BGDeployment` custom resource. If a change is detected, it will create the `Green` ReplicaSet `greenRS`. The operator will wait for a while for `greenRS` to become available. If `greenRS` manages to become available within the waiting period, the operator will modify label selectors of the service to match `greenRS`'s labels and scale down `blueRS` to zero replica (immediate rollback is possible as `blueRS` is scaled down but not deleted). Otherwise, the operator will scale down `greenRS` to zero replica and the service will still point to `blueRS`. During future rollouts, the operator will delete the ReplicaSet with zero replica to give way to new ReplicaSets. All these logics will be implemented via `UpdateFunc` event handler function. Finally, when the operator deletes the `BGDeployment` custom resource, the informer triggers immediate deletion of the two ReplicaSets and service via `DeleteFunc` event handler function.

The informer should have format below:

```
cache.NewInformer(
	crdclient.NewListWatch(), // client's ListWatch function
	&demov1.BGDeployment{}, // custom resource to watch
	1*time.Minute, // resync period
	cache.ResourceEventHandlerFuncs{ // event handler functions
		AddFunc: func(obj interface{}) {
			......
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			......
		},
		DeleteFunc: func(obj interface{}) {
			......
		},
	},
)
```

Refer this [article](https://engineering.bitnami.com/articles/a-deep-dive-into-kubernetes-controllers.html) for more information about controllers and informers.

## Custom Resource and Service

Steps below create `BGDeployment` custom resource and service.

1. Create a CRD
```
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: bgdeployments.demo.google.com
spec:
  group: demo.google.com
  version: v1
  scope: Namespaced
  names:
    plural: bgdeployments
    kind: BGDeployment
    shortNames:
    - bgd
```

2. Create a `BGDeployment` custom resource
```
apiVersion: demo.google.com/v1
kind: BGDeployment
metadata:
  name: blue-green-deployment
  namespace: default
  labels:
    app: nginx
spec:
  image: nginx:1.7.8
```

3. Create a service with label selectors matching the `Blue` deployment
```
kind: Service
apiVersion: core/v1
metadata:
  name: bgd-svc
  namespace: default
  labels:
    color: blue
spec:
  selector:
    color: blue
  ports:
  - protocol: TCP
    port: 80
    targetPort: 443
```

# Workflow

Workflow below demonstrates how the operator should behave.

1. Create `blueRS` with a valid nginx image (e.g., `nginx:1.7.8`), and wait for it to be available. When pinging the service, requests should reach `blueRS`'s pods successfully.

2. Update `.spec.image` field of the `BGDeployment` custom resource (e.g., update `nginx:1.7.8` to `nginx:1.7.9`). This triggers the operator to create `greenRS` with the newest image.

3. Wait for `greenRS` to be available.
    * If `greenRS` becomes available within the waiting period (i.e., image is valid), the operator updates label selectors of the service to match `greenRS`'s labels and scales down `blueRS` to zero replica. When pinging the service, requests should reach `greenRS`'s pods successfully.
    * Otherwise, the operator scales down `greenRS` to zero replica. `blueRS` and the service stay intact.

4. Repeat the steps above with another valid `.spec.image` field update (e.g., update `nginx:1.7.9` to `nginx:1.7.10`). You can see that the operator deletes the ReplicaSet with zero replica to make room for a new ReplicaSet with the newest image.

5. Repeat the steps above with an invalid `.spec.image` field update (e.g., update `nginx:1.7.10` to `nginx:1.7.100`). You can see that the operator first creates a new ReplicaSet, but scales the new ReplicaSet down after discovering the image is invalid.

Refer this [repository](https://github.com/crimsonfaith91/bgd-operator) for example codes. Feedback is welcome!
