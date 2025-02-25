/*
Copyright 2019 VMware

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"time"

	scheme "github.com/heptio/contour/apis/generated/clientset/versioned/scheme"
	v1alpha1 "github.com/heptio/contour/apis/projectcontour/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// HTTPProxiesGetter has a method to return a HTTPProxyInterface.
// A group's client should implement this interface.
type HTTPProxiesGetter interface {
	HTTPProxies(namespace string) HTTPProxyInterface
}

// HTTPProxyInterface has methods to work with HTTPProxy resources.
type HTTPProxyInterface interface {
	Create(*v1alpha1.HTTPProxy) (*v1alpha1.HTTPProxy, error)
	Update(*v1alpha1.HTTPProxy) (*v1alpha1.HTTPProxy, error)
	UpdateStatus(*v1alpha1.HTTPProxy) (*v1alpha1.HTTPProxy, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.HTTPProxy, error)
	List(opts v1.ListOptions) (*v1alpha1.HTTPProxyList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.HTTPProxy, err error)
	HTTPProxyExpansion
}

// hTTPProxies implements HTTPProxyInterface
type hTTPProxies struct {
	client rest.Interface
	ns     string
}

// newHTTPProxies returns a HTTPProxies
func newHTTPProxies(c *ProjectcontourV1alpha1Client, namespace string) *hTTPProxies {
	return &hTTPProxies{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the hTTPProxy, and returns the corresponding hTTPProxy object, and an error if there is any.
func (c *hTTPProxies) Get(name string, options v1.GetOptions) (result *v1alpha1.HTTPProxy, err error) {
	result = &v1alpha1.HTTPProxy{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("httpproxies").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of HTTPProxies that match those selectors.
func (c *hTTPProxies) List(opts v1.ListOptions) (result *v1alpha1.HTTPProxyList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.HTTPProxyList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("httpproxies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested hTTPProxies.
func (c *hTTPProxies) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("httpproxies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a hTTPProxy and creates it.  Returns the server's representation of the hTTPProxy, and an error, if there is any.
func (c *hTTPProxies) Create(hTTPProxy *v1alpha1.HTTPProxy) (result *v1alpha1.HTTPProxy, err error) {
	result = &v1alpha1.HTTPProxy{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("httpproxies").
		Body(hTTPProxy).
		Do().
		Into(result)
	return
}

// Update takes the representation of a hTTPProxy and updates it. Returns the server's representation of the hTTPProxy, and an error, if there is any.
func (c *hTTPProxies) Update(hTTPProxy *v1alpha1.HTTPProxy) (result *v1alpha1.HTTPProxy, err error) {
	result = &v1alpha1.HTTPProxy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("httpproxies").
		Name(hTTPProxy.Name).
		Body(hTTPProxy).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *hTTPProxies) UpdateStatus(hTTPProxy *v1alpha1.HTTPProxy) (result *v1alpha1.HTTPProxy, err error) {
	result = &v1alpha1.HTTPProxy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("httpproxies").
		Name(hTTPProxy.Name).
		SubResource("status").
		Body(hTTPProxy).
		Do().
		Into(result)
	return
}

// Delete takes name of the hTTPProxy and deletes it. Returns an error if one occurs.
func (c *hTTPProxies) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("httpproxies").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *hTTPProxies) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("httpproxies").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched hTTPProxy.
func (c *hTTPProxies) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.HTTPProxy, err error) {
	result = &v1alpha1.HTTPProxy{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("httpproxies").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
