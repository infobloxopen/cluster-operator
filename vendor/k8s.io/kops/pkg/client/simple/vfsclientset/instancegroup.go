/*
Copyright 2019 The Kubernetes Authors.

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

package vfsclientset

import (
	"fmt"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog"
	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/v1alpha1"
	"k8s.io/kops/pkg/apis/kops/validation"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/util/pkg/vfs"
)

type InstanceGroupVFS struct {
	commonVFS

	clusterName string
	cluster     *kopsapi.Cluster
}

type InstanceGroupMirror interface {
	WriteMirror(ig *kopsapi.InstanceGroup) error
}

var _ InstanceGroupMirror = &InstanceGroupVFS{}

func NewInstanceGroupMirror(cluster *kopsapi.Cluster, configBase vfs.Path) InstanceGroupMirror {
	if cluster == nil || cluster.Name == "" {
		klog.Fatalf("cluster / cluster.Name is required")
	}

	clusterName := cluster.Name
	kind := "InstanceGroup"

	r := &InstanceGroupVFS{
		cluster:     cluster,
		clusterName: clusterName,
	}
	r.init(kind, configBase.Join("instancegroup"), StoreVersion)
	defaultReadVersion := v1alpha1.SchemeGroupVersion.WithKind(kind)
	r.defaultReadVersion = &defaultReadVersion
	r.validate = func(o runtime.Object) error {
		return validation.ValidateInstanceGroup(o.(*kopsapi.InstanceGroup))
	}
	return r
}

func newInstanceGroupVFS(c *VFSClientset, cluster *kopsapi.Cluster) *InstanceGroupVFS {
	if cluster == nil || cluster.Name == "" {
		klog.Fatalf("cluster / cluster.Name is required")
	}

	clusterName := cluster.Name
	kind := "InstanceGroup"

	r := &InstanceGroupVFS{
		cluster:     cluster,
		clusterName: clusterName,
	}
	r.init(kind, c.basePath.Join(clusterName, "instancegroup"), StoreVersion)
	defaultReadVersion := v1alpha1.SchemeGroupVersion.WithKind(kind)
	r.defaultReadVersion = &defaultReadVersion
	r.validate = func(o runtime.Object) error {
		return validation.ValidateInstanceGroup(o.(*kopsapi.InstanceGroup))
	}
	return r
}

var _ kopsinternalversion.InstanceGroupInterface = &InstanceGroupVFS{}

func (c *InstanceGroupVFS) Get(name string, options metav1.GetOptions) (*kopsapi.InstanceGroup, error) {
	if options.ResourceVersion != "" {
		return nil, fmt.Errorf("ResourceVersion not supported in InstanceGroupVFS::Get")
	}

	o, err := c.find(name)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, errors.NewNotFound(schema.GroupResource{Group: kopsapi.GroupName, Resource: "InstanceGroup"}, name)
	}
	ig := o.(*kopsapi.InstanceGroup)
	c.addLabels(ig)

	return ig, nil
}

func (c *InstanceGroupVFS) addLabels(ig *kopsapi.InstanceGroup) {
	if ig.ObjectMeta.Labels == nil {
		ig.ObjectMeta.Labels = make(map[string]string)
	}
	ig.ObjectMeta.Labels[kopsapi.LabelClusterName] = c.clusterName
}

func (c *InstanceGroupVFS) List(options metav1.ListOptions) (*kopsapi.InstanceGroupList, error) {
	list := &kopsapi.InstanceGroupList{}
	items, err := c.list(list.Items, options)
	if err != nil {
		return nil, err
	}
	list.Items = items.([]kopsapi.InstanceGroup)
	for i := range list.Items {
		c.addLabels(&list.Items[i])
	}
	return list, nil
}

func (c *InstanceGroupVFS) Create(g *kopsapi.InstanceGroup) (*kopsapi.InstanceGroup, error) {
	err := c.create(c.cluster, g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (c *InstanceGroupVFS) Update(g *kopsapi.InstanceGroup) (*kopsapi.InstanceGroup, error) {

	old, err := c.Get(g.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if !apiequality.Semantic.DeepEqual(old.Spec, g.Spec) {
		g.SetGeneration(old.GetGeneration() + 1)
	}

	err = c.update(c.cluster, g)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (c *InstanceGroupVFS) WriteMirror(g *kopsapi.InstanceGroup) error {
	err := c.writeConfig(c.cluster, c.basePath.Join(g.Name), g)
	if err != nil {
		return fmt.Errorf("error writing %s: %v", c.kind, err)
	}

	return nil
}

func (c *InstanceGroupVFS) Delete(name string, options *metav1.DeleteOptions) error {
	return c.delete(name, options)
}

func (r *InstanceGroupVFS) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return fmt.Errorf("InstanceGroupVFS DeleteCollection not implemented for vfs store")
}

func (r *InstanceGroupVFS) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return nil, fmt.Errorf("InstanceGroupVFS Watch not implemented for vfs store")
}

func (r *InstanceGroupVFS) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *kopsapi.InstanceGroup, err error) {
	return nil, fmt.Errorf("InstanceGroupVFS Patch not implemented for vfs store")
}
