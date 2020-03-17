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

package openstacktasks

import (
	"fmt"

	"k8s.io/klog"
	// "github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/floatingips"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/listeners"
	"github.com/infobloxopen/cluster-operator/kops/upup/pkg/fi"
	"github.com/infobloxopen/cluster-operator/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=LBListener
type LBListener struct {
	ID        *string
	Name      *string
	Pool      *LBPool
	Lifecycle *fi.Lifecycle
}

// GetDependencies returns the dependencies of the Instance task
func (e *LBListener) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*LB); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*LBPool); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &LBListener{}

func (s *LBListener) CompareWithID() *string {
	return s.ID
}

func NewLBListenerTaskFromCloud(cloud openstack.OpenstackCloud, lifecycle *fi.Lifecycle, lb *listeners.Listener, find *LBListener) (*LBListener, error) {

	listenerTask := &LBListener{
		ID:        fi.String(lb.ID),
		Name:      fi.String(lb.Name),
		Lifecycle: lifecycle,
	}

	for _, pool := range lb.Pools {
		poolTask, err := NewLBPoolTaskFromCloud(cloud, lifecycle, &pool, find.Pool)
		if err != nil {
			return nil, fmt.Errorf("NewLBListenerTaskFromCloud: Failed to create new LBListener task for pool %s: %v", pool.Name, err)
		} else {
			listenerTask.Pool = poolTask
			// TODO: Support Multiple?
			break
		}
	}
	if find != nil {
		// Update all search terms
		find.ID = listenerTask.ID
		find.Name = listenerTask.Name
		find.Pool = listenerTask.Pool
	}
	return listenerTask, nil
}

func (s *LBListener) Find(context *fi.Context) (*LBListener, error) {
	if s.Name == nil {
		return nil, nil
	}

	cloud := context.Cloud.(openstack.OpenstackCloud)
	listenerList, err := cloud.ListListeners(listeners.ListOpts{
		ID:   fi.StringValue(s.ID),
		Name: fi.StringValue(s.Name),
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to list loadbalancer listeners for name %s: %v", fi.StringValue(s.Name), err)
	}
	if len(listenerList) == 0 {
		return nil, nil
	}
	if len(listenerList) > 1 {
		return nil, fmt.Errorf("Multiple listeners found with name %s", fi.StringValue(s.Name))
	}

	return NewLBListenerTaskFromCloud(cloud, s.Lifecycle, &listenerList[0], s)
}

func (s *LBListener) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, context)
}

func (_ *LBListener) CheckChanges(a, e, changes *LBListener) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *LBListener) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *LBListener) error {
	if a == nil {
		klog.V(2).Infof("Creating LB with Name: %q", fi.StringValue(e.Name))
		listeneropts := listeners.CreateOpts{
			Name:           fi.StringValue(e.Name),
			DefaultPoolID:  *e.Pool.ID,
			LoadbalancerID: *e.Pool.Loadbalancer.ID,
			Protocol:       listeners.ProtocolTCP,
			ProtocolPort:   443,
		}
		listener, err := t.Cloud.CreateListener(listeneropts)
		if err != nil {
			return fmt.Errorf("error creating LB listener: %v", err)
		}
		e.ID = fi.String(listener.ID)
		return nil
	}

	klog.V(2).Infof("Openstack task LB::RenderOpenstack did nothing")
	return nil
}
