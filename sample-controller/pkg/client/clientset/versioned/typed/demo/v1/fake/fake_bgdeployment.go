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

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	demo_v1 "k8s.io/sample-controller/pkg/apis/demo/v1"
)

// FakeBGDeployments implements BGDeploymentInterface
type FakeBGDeployments struct {
	Fake *FakeDemoV1
	ns   string
}

var bgdeploymentsResource = schema.GroupVersionResource{Group: "demo.crimsonfaith91.com", Version: "v1", Resource: "bgdeployments"}

var bgdeploymentsKind = schema.GroupVersionKind{Group: "demo.crimsonfaith91.com", Version: "v1", Kind: "BGDeployment"}

// Get takes name of the bGDeployment, and returns the corresponding bGDeployment object, and an error if there is any.
func (c *FakeBGDeployments) Get(name string, options v1.GetOptions) (result *demo_v1.BGDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(bgdeploymentsResource, c.ns, name), &demo_v1.BGDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*demo_v1.BGDeployment), err
}

// List takes label and field selectors, and returns the list of BGDeployments that match those selectors.
func (c *FakeBGDeployments) List(opts v1.ListOptions) (result *demo_v1.BGDeploymentList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(bgdeploymentsResource, bgdeploymentsKind, c.ns, opts), &demo_v1.BGDeploymentList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &demo_v1.BGDeploymentList{}
	for _, item := range obj.(*demo_v1.BGDeploymentList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested bGDeployments.
func (c *FakeBGDeployments) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(bgdeploymentsResource, c.ns, opts))

}

// Create takes the representation of a bGDeployment and creates it.  Returns the server's representation of the bGDeployment, and an error, if there is any.
func (c *FakeBGDeployments) Create(bGDeployment *demo_v1.BGDeployment) (result *demo_v1.BGDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(bgdeploymentsResource, c.ns, bGDeployment), &demo_v1.BGDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*demo_v1.BGDeployment), err
}

// Update takes the representation of a bGDeployment and updates it. Returns the server's representation of the bGDeployment, and an error, if there is any.
func (c *FakeBGDeployments) Update(bGDeployment *demo_v1.BGDeployment) (result *demo_v1.BGDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(bgdeploymentsResource, c.ns, bGDeployment), &demo_v1.BGDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*demo_v1.BGDeployment), err
}

// Delete takes name of the bGDeployment and deletes it. Returns an error if one occurs.
func (c *FakeBGDeployments) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(bgdeploymentsResource, c.ns, name), &demo_v1.BGDeployment{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBGDeployments) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(bgdeploymentsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &demo_v1.BGDeploymentList{})
	return err
}

// Patch applies the patch and returns the patched bGDeployment.
func (c *FakeBGDeployments) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *demo_v1.BGDeployment, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(bgdeploymentsResource, c.ns, name, data, subresources...), &demo_v1.BGDeployment{})

	if obj == nil {
		return nil, err
	}
	return obj.(*demo_v1.BGDeployment), err
}
