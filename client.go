/*
Copyright 2016 Iguazio Systems Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	demov1 "k8s.io/bgd-operator/pkg/apis/demo/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
)

func CrdClient(c kubernetes.Interface, cl *rest.RESTClient, scheme *runtime.Scheme, namespace string) *crdclient {
	return &crdclient{c: c, cl: cl, ns: namespace, plural: "bgdeployments",
		codec: runtime.NewParameterCodec(scheme)}
}

type crdclient struct {
	c      kubernetes.Interface
	cl     *rest.RESTClient
	ns     string
	plural string
	codec  runtime.ParameterCodec
}

func (f *crdclient) Create(obj *demov1.BGDeployment) (*demov1.BGDeployment, *extensionsv1beta1.ReplicaSet, error) {
	var result demov1.BGDeployment
	err := f.cl.Post().
		Namespace(f.ns).Resource(f.plural).
		Body(obj).Do().Into(&result)
	if err != nil {
		return &result, nil, err
	}

	// Create a RS along with CRD creation
	rs, err := f.CreateReplicaSet("blue-rs", "blue", obj)
	return &result, rs, err
}

func (f *crdclient) Update(obj *demov1.BGDeployment) (*demov1.BGDeployment, error) {
	var result demov1.BGDeployment
	err := f.cl.Put().
		Namespace(f.ns).Resource(f.plural).
		Body(obj).Do().Into(&result)
	return &result, err
}

func (f *crdclient) Delete(name string, options *metav1.DeleteOptions) error {
	return f.cl.Delete().
		Namespace(f.ns).Resource(f.plural).
		Name(name).Body(options).Do().
		Error()
}

func (f *crdclient) Get(name string) (*demov1.BGDeployment, error) {
	var result demov1.BGDeployment
	err := f.cl.Get().
		Namespace(f.ns).Resource(f.plural).
		Name(name).Do().Into(&result)
	return &result, err
}

func (f *crdclient) List(opts metav1.ListOptions) (*demov1.BGDeploymentList, error) {
	var result demov1.BGDeploymentList
	err := f.cl.Get().
		Namespace(f.ns).Resource(f.plural).
		VersionedParams(&opts, f.codec).
		Do().Into(&result)
	return &result, err
}

// Create a new List watch for custom resource
func (f *crdclient) NewListWatch() *cache.ListWatch {
	return cache.NewListWatchFromClient(f.cl, f.plural, f.ns, fields.Everything())
}

func newReplicaSet(name, color string, obj *demov1.BGDeployment) *extensionsv1beta1.ReplicaSet {
	one := int32(1)
	return &extensionsv1beta1.ReplicaSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ReplicaSet",
			APIVersion: "extensions/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: obj.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(obj, schema.GroupVersionKind{
					Group:   demov1.SchemeGroupVersion.Group,
					Version: demov1.SchemeGroupVersion.Version,
					Kind:    "BGDeployment",
				}),
			},
		},
		Spec: extensionsv1beta1.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"color": color},
			},
			Replicas: &one,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"color": color},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: obj.Spec.Image,
						},
					},
				},
			},
		},
	}
}

func (f *crdclient) CreateReplicaSet(name, color string, obj *demov1.BGDeployment) (*extensionsv1beta1.ReplicaSet, error) {
	return f.c.ExtensionsV1beta1().ReplicaSets(obj.Namespace).Create(newReplicaSet(name, color, obj))
}

func (f *crdclient) DeleteReplicaSet(rs *extensionsv1beta1.ReplicaSet) error {
	background := metav1.DeletePropagationBackground
	return f.c.ExtensionsV1beta1().ReplicaSets(rs.Namespace).Delete(rs.Name, &metav1.DeleteOptions{PropagationPolicy: &background})
}

func newService(namespace string) *corev1.Service {
	labels := map[string]string{"color": "blue"}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "core/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bgd-svc",
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       80,
					TargetPort: intstr.FromInt(443),
				},
			},
		},
	}
}

func (f *crdclient) CreateService(namespace string) (*corev1.Service, error) {
	return f.c.CoreV1().Services(namespace).Create(newService(namespace))
}

func (f *crdclient) UpdateService(svcName, namespace string, updateFunc func(*corev1.Service)) (*corev1.Service, error) {
	var svc *corev1.Service
	svcClient := f.c.CoreV1().Services(namespace)
	if err := retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		newSvc, err := svcClient.Get(svcName, metav1.GetOptions{})
		if err != nil {
			return err
		}
		updateFunc(newSvc)
		svc, err = svcClient.Update(newSvc)
		return err
	}); err != nil {
		return nil, err
	}
	return svc, nil
}

func (f *crdclient) DeleteService(svc *corev1.Service) error {
	return f.c.CoreV1().Services(svc.Namespace).Delete(svc.Name, &metav1.DeleteOptions{})
}

// waitAllPodsAvailable returns true if all pods are available, false otherwise
func (f *crdclient) WaitAllPodsAvailable(rs *extensionsv1beta1.ReplicaSet, pollInterval, pollTimeout time.Duration) bool {
	if err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		newRS, err := f.c.ExtensionsV1beta1().ReplicaSets(rs.Namespace).Get(rs.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return newRS.Status.Replicas == *rs.Spec.Replicas && newRS.Status.AvailableReplicas == *rs.Spec.Replicas, nil
	}); err != nil {
		fmt.Printf("failed to wait for all pods of replicaset %q to be available: %v\n", rs.Name, err)
		return false
	}
	return true
}

func NewClient(cfg *rest.Config) (*rest.RESTClient, *runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	SchemeBuilder := runtime.NewSchemeBuilder(demov1.AddKnownTypes)
	if err := SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, nil, err
	}
	config := *cfg
	config.GroupVersion = &demov1.SchemeGroupVersion
	config.APIPath = "/apis"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{
		CodecFactory: serializer.NewCodecFactory(scheme),
	}

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, nil, err
	}
	return client, scheme, nil
}
