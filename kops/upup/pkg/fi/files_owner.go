// +build !windows

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

package fi

import (
	"fmt"
	"os"
	"syscall"

	"k8s.io/klog"
)

func EnsureFileOwner(destPath string, owner string, groupName string) (bool, error) {
	changed := false
	stat, err := os.Lstat(destPath)
	if err != nil {
		return changed, fmt.Errorf("error getting file stat for %q: %v", destPath, err)
	}

	actualUserID := int(stat.Sys().(*syscall.Stat_t).Uid)
	userID := actualUserID
	if owner != "" {
		user, err := LookupUser(owner) //user.Lookup(owner)
		if err != nil {
			return changed, fmt.Errorf("error looking up user %q: %v", owner, err)
		}
		if user == nil {
			return changed, fmt.Errorf("user %q not found", owner)
		}
		userID = user.Uid
	}

	actualGroupID := int(stat.Sys().(*syscall.Stat_t).Gid)
	groupID := actualGroupID
	if groupName != "" {
		group, err := LookupGroup(groupName)
		if err != nil {
			return changed, fmt.Errorf("error looking up group %q: %v", groupName, err)
		}
		if group == nil {
			return changed, fmt.Errorf("group %q not found", groupName)
		}
		groupID = group.Gid
	}

	if actualUserID == userID && actualGroupID == groupID {
		return changed, nil
	}

	klog.Infof("Changing file owner/group for %q to %s:%s", destPath, owner, groupName)
	err = os.Lchown(destPath, userID, groupID)
	if err != nil {
		return changed, fmt.Errorf("error setting file owner/group for %q: %v", destPath, err)
	}
	changed = true

	return changed, nil
}
