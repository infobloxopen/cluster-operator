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

package local

import (
	"os/exec"

	"k8s.io/apimachinery/pkg/util/sets"
	"github.com/infobloxopen/cluster-operator/kops/upup/pkg/fi"
)

type LocalTarget struct {
	CacheDir string
	Tags     sets.String
}

var _ fi.Target = &LocalTarget{}

func (t *LocalTarget) Finish(taskMap map[string]fi.Task) error {
	return nil
}

func (t *LocalTarget) ProcessDeletions() bool {
	// We don't expect any, but it would be our job to process them
	return true
}

func (t *LocalTarget) HasTag(tag string) bool {
	_, found := t.Tags[tag]
	return found
}

// CombinedOutput is a helper function that executes a command, returning stdout & stderr combined
func (t *LocalTarget) CombinedOutput(args []string) ([]byte, error) {
	c := exec.Command(args[0], args[1:]...)
	return c.CombinedOutput()
}
