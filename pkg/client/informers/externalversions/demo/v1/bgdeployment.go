/*
Copyright 2017 The Kubernetes Authors.

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

// This file was automatically generated by informer-gen

package v1

import (
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	demo_v1 "k8s.io/bgd-operator/pkg/apis/demo/v1"
	versioned "k8s.io/bgd-operator/pkg/client/clientset/versioned"
	internalinterfaces "k8s.io/bgd-operator/pkg/client/informers/externalversions/internalinterfaces"
	v1 "k8s.io/bgd-operator/pkg/client/listers/demo/v1"
	cache "k8s.io/client-go/tools/cache"
	time "time"
)

// BGDeploymentInformer provides access to a shared informer and lister for
// BGDeployments.
type BGDeploymentInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.BGDeploymentLister
}

type bGDeploymentInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewBGDeploymentInformer constructs a new informer for BGDeployment type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewBGDeploymentInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredBGDeploymentInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredBGDeploymentInformer constructs a new informer for BGDeployment type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredBGDeploymentInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DemoV1().BGDeployments(namespace).List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.DemoV1().BGDeployments(namespace).Watch(options)
			},
		},
		&demo_v1.BGDeployment{},
		resyncPeriod,
		indexers,
	)
}

func (f *bGDeploymentInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredBGDeploymentInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *bGDeploymentInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&demo_v1.BGDeployment{}, f.defaultInformer)
}

func (f *bGDeploymentInformer) Lister() v1.BGDeploymentLister {
	return v1.NewBGDeploymentLister(f.Informer().GetIndexer())
}
