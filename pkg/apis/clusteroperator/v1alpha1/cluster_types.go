package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KopsCluster defines the settings passed to Kops
// +k8s:openapi-gen=true
type KopsConfig struct {
	Name        string   `json:"name,omitempty"`
	MasterCount int      `json:"master_count,omitempty"`
	MasterEc2   string   `json:"master_ec2,omitempty"`
	WorkerCount int      `json:"worker_count,omitempty"`
	WorkerEc2   string   `json:"worker_ec2,omitempty"`
	StateStore  string   `json:"state_store,omitempty"`
	Vpc         string   `json:"vpc,omitempty"`
	Zones       []string `json:"zones,omitempty"`
}

// KopsFailure informs regarding reason cluster is not ready
// +k8s:openapi-gen=true
type KopsFailure struct {
	Type    string `json:"type,omitempty"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message,omitempty"`
}

// KopsNodes resports about the cluster nodes when ready
// +k8s:openapi-gen=true
type KopsNodes struct {
	Name     string `json:"name,omitempty"`
	Zone     string `json:"zone,omitempty"`
	Role     string `json:"role,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Status   string `json:"status,omitempty"`
}

// KubeConfig holds the config to access the cluster
// +k8s:openapi-gen=true
type KubeConfig struct {
	APIVersion     string           `json:"apiVersion"`
	Clusters       []ClusterConfigs `json:"clusters"`
	Contexts       []Contexts       `json:"contexts"`
	CurrentContext string           `json:"current-context"`
	Kind           string           `json:"kind"`
	// Preferences    struct {
	// } `yaml:"preferences"`
	Users []Users `json:"users"`
}

// ClusterConfig defines attributes for kubeconfig cluster
// +k8s:openapi-gen=true
type ClusterConfig struct {
	CertificateAuthorityData string `json:"certificate-authority-data"`
	Server                   string `json:"server"`
}

// ClusterConfigs defines a collection of cluster configs
// +k8s:openapi-gen=true
type ClusterConfigs struct {
	ClusterConfigs []ClusterConfigs
	Name           string `json:"name"`
}

// Context defines attributes for kubeconfig
// +k8s:openapi-gen=true
type Context struct {
	Cluster string `json:"cluster"`
	User    string `json:"user"`
}

// Contexts defines a list of contexts for a kubeconfig
// +k8s:openapi-gen=true
type Contexts struct {
	Context []Context `json:"context,omitempty"`
	Name    string    `json:"name"`
}

// User defines a user for a cluster in a kubeconfig
// +k8s:openapi-gen=true
type User struct {
	ClientCertificateData string `json:"client-certificate-data"`
	ClientKeyData         string `json:"client-key-data"`
}

// Users is a list of 'User' that defines acces through a kubeconfig
// +k8s:openapi-gen=true
type Users struct {
	Users []User `json:"users,omitempty"`
	Name  string `json:"name,omitempty"`
}

// KopsStatus defines the status of the Kops Cluster
// +k8s:openapi-gen=true
type KopsStatus struct {
	Failures []KopsFailure `json:"failures,omitempty"`
	Nodes    []KopsNodes   `json:"nodes,omitempty"`
}

// ClusterSpec defines the desired state of Cluster
// +k8s:openapi-gen=true
type ClusterSpec struct {
	// Name is the desired k8s cluster name
	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	Name string `json:"name,omitempty"`
	// Kops Cluster Config
	KopsConfig KopsConfig `json:"kops_config,omitempty"`
}

// PodPhase is a label for the condition of a pod at the current time.
type ClusterPhase string

// These are the valid statuses of pods.
const (
	// PodPending means the Cluster has not been configured yet. We configure
	// desired state in state store and transistion to ClusterUpdate
	ClusterPending ClusterPhase = "Pending"
	// ClusterUpdate means the Cluster has been configured by the system, but cluster
	// has not been completely been configured
	ClusterUpdate ClusterPhase = "Update"
	// Setup is set when we are waiting for  the Cluster to come up
	// so that it can be used.
	ClusterSetup ClusterPhase = "Setup"
	// ClusterDone means that Cluster has been provisioned
	// and can be used
	ClusterDone ClusterPhase = "Done"
)

// ClusterStatus defines the observed state of Cluster
// +k8s:openapi-gen=true
type ClusterStatus struct {
	// Phase represents the state of the cluster provisioning
	// It transitions from PENDING to DONE, we might add more states for infrastructure provisioning
	Phase ClusterPhase `json:"phase,omitempty"`
	// Kops Cluster Status
	KopsStatus KopsStatus `json:"kops_status,omitempty"`
	KubeConfig KubeConfig `json:"kubeconfig,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cluster is the Schema for the clusters API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=clusters,scope=Namespaced
// +k8s:openapi-gen=true
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
