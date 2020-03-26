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

/******************************************************************************
* The Kops Tag Builder
*
* Tags are how we manage kops functionality.
*
******************************************************************************/

package cloudup

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func buildCloudupTags(cluster *api.Cluster) (sets.String, error) {
	tags := sets.NewString()

	switch api.CloudProviderID(cluster.Spec.CloudProvider) {
	case api.CloudProviderGCE:
		{
			tags.Insert("_gce")
		}

	case api.CloudProviderAWS:
		{
			tags.Insert("_aws")
		}
	case api.CloudProviderDO:
		{
			tags.Insert("_do")
		}
	case api.CloudProviderVSphere:
		{
			tags.Insert("_vsphere")
		}

	case api.CloudProviderBareMetal:
		// No tags

	case api.CloudProviderOpenstack:

	case api.CloudProviderALI:
		{
			tags.Insert("_ali")
		}
	default:
		return nil, fmt.Errorf("unknown CloudProvider %q", cluster.Spec.CloudProvider)
	}

	tags.Insert("_k8s_1_6")

	klog.V(4).Infof("tags: %s", tags.List())

	return tags, nil
}

func buildNodeupTags(role api.InstanceGroupRole, cluster *api.Cluster, clusterTags sets.String) (sets.String, error) {
	tags := sets.NewString()

	networking := cluster.Spec.Networking

	if networking == nil {
		return nil, fmt.Errorf("Networking is not set, and should not be nil here")
	}

	if networking.LyftVPC != nil {
		tags.Insert("_lyft_vpc_cni")
	}

	switch fi.StringValue(cluster.Spec.UpdatePolicy) {
	case "": // default
		tags.Insert("_automatic_upgrades")
	case api.UpdatePolicyExternal:
	// Skip applying the tag
	default:
		klog.Warningf("Unrecognized value for UpdatePolicy: %v", fi.StringValue(cluster.Spec.UpdatePolicy))
	}

	if clusterTags.Has("_gce") {
		tags.Insert("_gce")
	}
	if clusterTags.Has("_aws") {
		tags.Insert("_aws")
	}
	if clusterTags.Has("_do") {
		tags.Insert("_do")
	}

	return tags, nil
}
