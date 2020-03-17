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

package nodetasks

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"k8s.io/klog"
	"github.com/infobloxopen/cluster-operator/kops/upup/pkg/fi"
	"github.com/infobloxopen/cluster-operator/kops/upup/pkg/fi/nodeup/cloudinit"
	"github.com/infobloxopen/cluster-operator/kops/upup/pkg/fi/nodeup/local"
	"github.com/infobloxopen/cluster-operator/kops/upup/pkg/fi/nodeup/tags"
)

type UpdatePackages struct {
	// We can't be completely empty or we don't run
	Updated bool
}

var _ fi.HasDependencies = &UpdatePackages{}

func NewUpdatePackages() *UpdatePackages {
	return &UpdatePackages{Updated: true}
}

func (p *UpdatePackages) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return []fi.Task{}
}

func (p *UpdatePackages) String() string {
	return fmt.Sprintf("UpdatePackages")
}

func (e *UpdatePackages) Find(c *fi.Context) (*UpdatePackages, error) {
	return nil, nil
}

func (e *UpdatePackages) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (s *UpdatePackages) CheckChanges(a, e, changes *UpdatePackages) error {
	return nil
}

func (_ *UpdatePackages) RenderLocal(t *local.LocalTarget, a, e, changes *UpdatePackages) error {
	if os.Getenv("SKIP_PACKAGE_UPDATE") != "" {
		klog.Infof("SKIP_PACKAGE_UPDATE was set; skipping package update")
		return nil
	}
	var args []string
	if t.HasTag(tags.TagOSFamilyDebian) {
		args = []string{"apt-get", "update"}

	} else if t.HasTag(tags.TagOSFamilyRHEL) {
		// Probably not technically needed
		args = []string{"/usr/bin/yum", "check-update"}
	} else {
		return fmt.Errorf("unsupported package system")
	}
	klog.Infof("running command %s", args)
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	// 'yum check-update' exits with 100 if it finds updates; treat it like a success
	if exitCode := cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus(); err != nil && exitCode != 100 {
		return fmt.Errorf("error update packages: %v: %s", err, string(output))
	}

	return nil
}

func (_ *UpdatePackages) RenderCloudInit(t *cloudinit.CloudInitTarget, a, e, changes *UpdatePackages) error {
	t.Config.PackageUpdate = true
	return nil
}
