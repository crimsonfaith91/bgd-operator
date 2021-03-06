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

	"flag"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	demov1 "k8s.io/bgd-operator/pkg/apis/demo/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClientConfig returns rest config, if path not specified assume in cluster config
func GetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func main() {
	kubeconf := flag.String("kubeconf", "admin.conf", "Path to a kube config. Only required if out-of-cluster.")
	flag.Parse()

	config, err := GetClientConfig(*kubeconf)
	if err != nil {
		panic(err.Error())
	}

	// Create a new clientset which includes the CRD schema
	crdcs, scheme, err := NewClient(config)
	if err != nil {
		panic(err)
	}

	// Create a CRD client interface
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("Error building kubernetes clientset: %s", err.Error()))
	}

	crdclient := CrdClient(kubeClient, crdcs, scheme, "default")
	image := "nginx:1.7.9"
	colorMap := map[string]string{"blue": "green", "green": "blue"}
	var rs *extensionsv1beta1.ReplicaSet
	var svc *corev1.Service

	// Create an informer that watches changes in BGDeployment custom resource
	_, controller := cache.NewInformer(
		crdclient.NewListWatch(),
		&demov1.BGDeployment{},
		1*time.Minute,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				fmt.Printf("Add: %+v\n", obj)
				bgd := obj.(*demov1.BGDeployment)

				// Create a blue RS along with CRD creation
				rs, err = crdclient.CreateReplicaSet("blue-rs", "blue", bgd)
				if err == nil {
					fmt.Printf("created replicaset %q\n", rs.Name)
				} else if apierrors.IsAlreadyExists(err) {
					fmt.Printf("replicaset already exists")
				} else {
					panic(err)
				}

				// Create a service along with CRD creation
				svc, err = crdclient.CreateService(bgd.Namespace)
				if err != nil {
					panic(fmt.Sprintf("failed to create service: %v", err))
				}
			},
			DeleteFunc: func(obj interface{}) {
				fmt.Printf("Delete: %+v\n", obj)

				// Delete service when the BGDeployment custom resource is deleted
				err = crdclient.DeleteService(svc)
				if err != nil {
					panic(fmt.Sprintf("failed to delete service when the BGDeployment custom resource is deleted: %v", err))
				}
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				fmt.Printf("Update Old: %+v\n\nNew: %+v\n", oldObj, newObj)
				bgd := newObj.(*demov1.BGDeployment)
				newImage := bgd.Spec.Image

				// Only create the new RS when the image is changed
				if newImage != image {
					image = newImage
					var newColor string

					// Before creating another RS, look for RS with zero replica
					// If the RS with zero replica exists, update new color with its color and delete it
					// Else, update the new color with another color
					rss, err := crdclient.ListReplicaSet(bgd.Namespace)
					if err != nil {
						panic(fmt.Sprintf("failed to list RSs: %v", err))
					}
					for _, curRS := range rss.Items {
						selectorMap, err := metav1.LabelSelectorAsMap(curRS.Spec.Selector)
						if err != nil {
							panic(fmt.Sprintf("failed to convert label selector of RS %q to a map: %v", curRS.Name, err))
						}
						curColor := selectorMap["color"]

						if curRS.Status.AvailableReplicas == 0 {
							// Update the new color with color of the RS with zero replica
							newColor = curColor

							// Delete the RS with zero replica
							err = crdclient.DeleteReplicaSet(&curRS)
							if err != nil {
								panic(fmt.Sprintf("failed to delete RS %q with zero replica before creating a new RS with newest image name: %v", curRS.Name, err))
							}

							// Immediate break out for the loop to prevent the new color from being updated again
							break
						} else {
							// Update the new color with another color
							newColor = colorMap[curColor]
						}
					}

					// Create a new RS with the new color
					newRS, err := crdclient.CreateReplicaSet(fmt.Sprintf("%s-rs", newColor), newColor, bgd)
					if err != nil {
						panic(fmt.Sprintf("failed to create new RS when image is changed: %v", err))
					}

					// Determine whether all pods of the new RS are available (i.e., ready)
					allNewPodsAvailable := crdclient.WaitAllPodsAvailable(newRS, 100*time.Millisecond, 5*time.Second)
					if allNewPodsAvailable {
						// Update service to point to the new RS
						svc, err = crdclient.UpdateService(svc.Name, bgd.Namespace, func(service *corev1.Service) {
							updatedLabels := map[string]string{"color": newColor}
							service.Labels = updatedLabels
							service.Spec.Selector = updatedLabels
						})
						if err != nil {
							panic(fmt.Sprintf("failed to update service to point to new RS, %q: %v", newRS.Name, err))
						}

						// Scale down the old RS to zero replica
						err = crdclient.ScaleReplicaSet(rs, 0)
						if err != nil {
							panic(fmt.Sprintf("failed to scale down old RS to zero replica: %v", err))
						}

						// Change rs variable to point to newRS
						rs = newRS
					} else {
						// Scale down the new RS to zero replica
						err = crdclient.ScaleReplicaSet(newRS, 0)
						if err != nil {
							panic(fmt.Sprintf("failed to scale down new RS to zero replica: %v", err))
						}
					}
				}
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	// Wait forever to ensure BGDeployment controller is running indefinitely
	select {}
}
