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

package simple

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	kopsinternalversion "k8s.io/kops/pkg/client/clientset_generated/clientset/typed/kops/internalversion"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

type Clientset interface {
	// GetCluster reads a cluster by name
	GetCluster(name string) (*kops.Cluster, error)

	// CreateCluster creates a cluster
	CreateCluster(cluster *kops.Cluster) (*kops.Cluster, error)

	// UpdateCluster updates a cluster
	UpdateCluster(cluster *kops.Cluster, status *kops.ClusterStatus) (*kops.Cluster, error)

	// ListClusters returns all clusters
	ListClusters(options metav1.ListOptions) (*kops.ClusterList, error)

	// ConfigBaseFor returns the vfs path where we will read configuration information from
	ConfigBaseFor(cluster *kops.Cluster) (vfs.Path, error)

	// InstanceGroupsFor returns the InstanceGroupInterface bounds to the namespace for a particular Cluster
	InstanceGroupsFor(cluster *kops.Cluster) kopsinternalversion.InstanceGroupInterface

	// SecretStore builds the secret store for the specified cluster
	SecretStore(cluster *kops.Cluster) (fi.SecretStore, error)

	// KeyStore builds the key store for the specified cluster
	KeyStore(cluster *kops.Cluster) (fi.CAStore, error)

	// SSHCredentialStore builds the SSHCredential store for the specified cluster
	SSHCredentialStore(cluster *kops.Cluster) (fi.SSHCredentialStore, error)

	// DeleteCluster deletes all the state for the specified cluster
	DeleteCluster(cluster *kops.Cluster) error
}
