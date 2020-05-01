package clustervalidator

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	// Mock Admission Request for a new Cluster resource
	AdmissionRequestCreate = v1beta1.AdmissionReview{
		TypeMeta: v1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &v1beta1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: v1.GroupVersionKind{
				Group:   "cluster-operator.infobloxopen.github.com",
				Version: "v1alpha1",
				Kind:    "Cluster",
			},
			Operation: "CREATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "cluster-operator.infobloxopen.github.com/v1alpha1",
					"kind": "Cluster",
					"metadata": {
						"name": "example-cluster",
						"namespace": "scoleman",
						"resourceVersion": "16932",
						"selfLink": "/apis/cluster-operator.infobloxopen.github.com/v1alpha1/namespaces/scoleman/clusters/example-cluster",
						"uid": "5516629b-b095-4975-a316-3339002855b0"
					},
					"spec": {
						"config": "apiVersion: kops.k8s.io/v1alpha2\nkind: Cluster\nmetadata:\n  name: scoleman.soheil.belamaric.com\nspec:\n  api:\n    dns: {}\n  authorization:\n    rbac: {}\n  channel: stable\n  cloudProvider: aws\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\n  etcdClusters:\n  - cpuRequest: 200m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: main\n  - cpuRequest: 100m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: events\n  iam:\n    allowContainerRegistry: true\n    legacy: false\n  kubelet:\n    anonymousAuth: false\n  kubernetesApiAccess:\n  - 0.0.0.0/0\n  kubernetesVersion: 1.16.7\n  masterPublicName: api.scoleman.soheil.belamaric.com\n  networkCIDR: 172.17.16.0/21\n  networkID: vpc-0a75b33895655b46a\n  networking:\n    kubenet: {}\n  nonMasqueradeCIDR: 100.64.0.0/10\n  sshAccess:\n  - 0.0.0.0/0\n  subnets:\n  - cidr: 172.17.17.0/24\n    name: us-east-2a\n    type: Public\n    zone: us-east-2a\n  - cidr: 172.17.18.0/24\n    name: us-east-2b\n    type: Public\n    zone: us-east-2b\n  topology:\n    dns:\n      type: Public\n    masters: public\n    nodes: public\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: master-us-east-2a\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 1\n  minSize: 1\n  nodeLabels:\n    kops.k8s.io/instancegroup: master-us-east-2a\n  role: Master\n  subnets:\n  - us-east-2a\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: nodes\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 2\n  minSize: 2\n  nodeLabels:\n    kops.k8s.io/instancegroup: nodes\n  role: Node\n  subnets:\n  - us-east-2a\n  - us-east-2b\n---\napiVersion: kops/v1alpha2\nkind: SSHCredential\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\nspec:\n  publicKey: \"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\"\n",
						"name": "scoleman"
					}
				}`),
			},
		},
	}
	// Mock Admission Request for a updating a Cluster resource without changing Spec.Name
	// Expected to be allowed by the validating webhook
	AdmissionRequestUpdateSameName = v1beta1.AdmissionReview{
		TypeMeta: v1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &v1beta1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: v1.GroupVersionKind{
				Group:   "cluster-operator.infobloxopen.github.com",
				Version: "v1alpha1",
				Kind:    "Cluster",
			},
			Operation: "UPDATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "cluster-operator.infobloxopen.github.com/v1alpha1",
					"kind": "Cluster",
					"metadata": {
						"annotations": {
							"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"cluster-operator.infobloxopen.github.com/v1alpha1\",\"kind\":\"Cluster\",\"metadata\":{\"annotations\":{},\"name\":\"example-cluster\",\"namespace\":\"scoleman\"},\"spec\":{\"config\":\"apiVersion: kops.k8s.io/v1alpha2\\nkind: Cluster\\nmetadata:\\n  name: scoleman.soheil.belamaric.com\\nspec:\\n  api:\\n    dns: {}\\n  authorization:\\n    rbac: {}\\n  channel: stable\\n  cloudProvider: aws\\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\\n  etcdClusters:\\n  - cpuRequest: 200m\\n    etcdMembers:\\n    - instanceGroup: master-us-east-2a\\n      name: a\\n    memoryRequest: 100Mi\\n    name: main\\n  - cpuRequest: 100m\\n    etcdMembers:\\n    - instanceGroup: master-us-east-2a\\n      name: a\\n    memoryRequest: 100Mi\\n    name: events\\n  iam:\\n    allowContainerRegistry: true\\n    legacy: false\\n  kubelet:\\n    anonymousAuth: false\\n  kubernetesApiAccess:\\n  - 0.0.0.0/0\\n  kubernetesVersion: 1.16.7\\n  masterPublicName: api.scoleman.soheil.belamaric.com\\n  networkCIDR: 172.17.16.0/21\\n  networkID: vpc-0a75b33895655b46a\\n  networking:\\n    kubenet: {}\\n  nonMasqueradeCIDR: 100.64.0.0/10\\n  sshAccess:\\n  - 0.0.0.0/0\\n  subnets:\\n  - cidr: 172.17.17.0/24\\n    name: us-east-2a\\n    type: Public\\n    zone: us-east-2a\\n  - cidr: 172.17.18.0/24\\n    name: us-east-2b\\n    type: Public\\n    zone: us-east-2b\\n  topology:\\n    dns:\\n      type: Public\\n    masters: public\\n    nodes: public\\n---\\napiVersion: kops.k8s.io/v1alpha2\\nkind: InstanceGroup\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\n  name: master-us-east-2a\\nspec:\\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\\n  machineType: t2.micro\\n  maxSize: 1\\n  minSize: 1\\n  nodeLabels:\\n    kops.k8s.io/instancegroup: master-us-east-2a\\n  role: Master\\n  subnets:\\n  - us-east-2a\\n---\\napiVersion: kops.k8s.io/v1alpha2\\nkind: InstanceGroup\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\n  name: nodes\\nspec:\\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\\n  machineType: t2.micro\\n  maxSize: 2\\n  minSize: 2\\n  nodeLabels:\\n    kops.k8s.io/instancegroup: nodes\\n  role: Node\\n  subnets:\\n  - us-east-2a\\n  - us-east-2b\\n---\\napiVersion: kops/v1alpha2\\nkind: SSHCredential\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\nspec:\\n  publicKey: \\\"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\\\"\\n\",\"name\":\"scoleman-3\"}}\n"
						},
						"creationTimestamp": "2020-04-22T22:01:40Z",
						"generation": 3,
						"name": "example-cluster",
						"namespace": "scoleman",
						"resourceVersion": "16932",
						"selfLink": "/apis/cluster-operator.infobloxopen.github.com/v1alpha1/namespaces/scoleman/clusters/example-cluster",
						"uid": "5516629b-b095-4975-a316-3339002855b0"
					},
					"spec": {
						"config": "apiVersion: kops.k8s.io/v1alpha2\nkind: Cluster\nmetadata:\n  name: scoleman.soheil.belamaric.com\nspec:\n  api:\n    dns: {}\n  authorization:\n    rbac: {}\n  channel: stable\n  cloudProvider: aws\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\n  etcdClusters:\n  - cpuRequest: 200m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: main\n  - cpuRequest: 100m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: events\n  iam:\n    allowContainerRegistry: true\n    legacy: false\n  kubelet:\n    anonymousAuth: false\n  kubernetesApiAccess:\n  - 0.0.0.0/0\n  kubernetesVersion: 1.16.7\n  masterPublicName: api.scoleman.soheil.belamaric.com\n  networkCIDR: 172.17.16.0/21\n  networkID: vpc-0a75b33895655b46a\n  networking:\n    kubenet: {}\n  nonMasqueradeCIDR: 100.64.0.0/10\n  sshAccess:\n  - 0.0.0.0/0\n  subnets:\n  - cidr: 172.17.17.0/24\n    name: us-east-2a\n    type: Public\n    zone: us-east-2a\n  - cidr: 172.17.18.0/24\n    name: us-east-2b\n    type: Public\n    zone: us-east-2b\n  topology:\n    dns:\n      type: Public\n    masters: public\n    nodes: public\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: master-us-east-2a\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 1\n  minSize: 1\n  nodeLabels:\n    kops.k8s.io/instancegroup: master-us-east-2a\n  role: Master\n  subnets:\n  - us-east-2a\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: nodes\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 2\n  minSize: 2\n  nodeLabels:\n    kops.k8s.io/instancegroup: nodes\n  role: Node\n  subnets:\n  - us-east-2a\n  - us-east-2b\n---\napiVersion: kops/v1alpha2\nkind: SSHCredential\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\nspec:\n  publicKey: \"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\"\n",
						"name": "scoleman"
					}
				}`),
			},
			OldObject: runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "cluster-operator.infobloxopen.github.com/v1alpha1",
					"kind": "Cluster",
					"metadata": {
						"annotations": {
							"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"cluster-operator.infobloxopen.github.com/v1alpha1\",\"kind\":\"Cluster\",\"metadata\":{\"annotations\":{},\"name\":\"example-cluster\",\"namespace\":\"scoleman\"},\"spec\":{\"config\":\"apiVersion: kops.k8s.io/v1alpha2\\nkind: Cluster\\nmetadata:\\n  name: scoleman.soheil.belamaric.com\\nspec:\\n  api:\\n    dns: {}\\n  authorization:\\n    rbac: {}\\n  channel: stable\\n  cloudProvider: aws\\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\\n  etcdClusters:\\n  - cpuRequest: 200m\\n    etcdMembers:\\n    - instanceGroup: master-us-east-2a\\n      name: a\\n    memoryRequest: 100Mi\\n    name: main\\n  - cpuRequest: 100m\\n    etcdMembers:\\n    - instanceGroup: master-us-east-2a\\n      name: a\\n    memoryRequest: 100Mi\\n    name: events\\n  iam:\\n    allowContainerRegistry: true\\n    legacy: false\\n  kubelet:\\n    anonymousAuth: false\\n  kubernetesApiAccess:\\n  - 0.0.0.0/0\\n  kubernetesVersion: 1.16.7\\n  masterPublicName: api.scoleman.soheil.belamaric.com\\n  networkCIDR: 172.17.16.0/21\\n  networkID: vpc-0a75b33895655b46a\\n  networking:\\n    kubenet: {}\\n  nonMasqueradeCIDR: 100.64.0.0/10\\n  sshAccess:\\n  - 0.0.0.0/0\\n  subnets:\\n  - cidr: 172.17.17.0/24\\n    name: us-east-2a\\n    type: Public\\n    zone: us-east-2a\\n  - cidr: 172.17.18.0/24\\n    name: us-east-2b\\n    type: Public\\n    zone: us-east-2b\\n  topology:\\n    dns:\\n      type: Public\\n    masters: public\\n    nodes: public\\n---\\napiVersion: kops.k8s.io/v1alpha2\\nkind: InstanceGroup\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\n  name: master-us-east-2a\\nspec:\\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\\n  machineType: t2.micro\\n  maxSize: 1\\n  minSize: 1\\n  nodeLabels:\\n    kops.k8s.io/instancegroup: master-us-east-2a\\n  role: Master\\n  subnets:\\n  - us-east-2a\\n---\\napiVersion: kops.k8s.io/v1alpha2\\nkind: InstanceGroup\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\n  name: nodes\\nspec:\\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\\n  machineType: t2.micro\\n  maxSize: 2\\n  minSize: 2\\n  nodeLabels:\\n    kops.k8s.io/instancegroup: nodes\\n  role: Node\\n  subnets:\\n  - us-east-2a\\n  - us-east-2b\\n---\\napiVersion: kops/v1alpha2\\nkind: SSHCredential\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\nspec:\\n  publicKey: \\\"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\\\"\\n\",\"name\":\"scoleman-3\"}}\n"
						},
						"creationTimestamp": "2020-04-22T22:01:40Z",
						"generation": 3,
						"name": "example-cluster",
						"namespace": "scoleman",
						"resourceVersion": "16932",
						"selfLink": "/apis/cluster-operator.infobloxopen.github.com/v1alpha1/namespaces/scoleman/clusters/example-cluster",
						"uid": "5516629b-b095-4975-a316-3339002855b0"
					},
					"spec": {
						"config": "apiVersion: kops.k8s.io/v1alpha2\nkind: Cluster\nmetadata:\n  name: scoleman.soheil.belamaric.com\nspec:\n  api:\n    dns: {}\n  authorization:\n    rbac: {}\n  channel: stable\n  cloudProvider: aws\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\n  etcdClusters:\n  - cpuRequest: 200m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: main\n  - cpuRequest: 100m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: events\n  iam:\n    allowContainerRegistry: true\n    legacy: false\n  kubelet:\n    anonymousAuth: false\n  kubernetesApiAccess:\n  - 0.0.0.0/0\n  kubernetesVersion: 1.16.7\n  masterPublicName: api.scoleman.soheil.belamaric.com\n  networkCIDR: 172.17.16.0/21\n  networkID: vpc-0a75b33895655b46a\n  networking:\n    kubenet: {}\n  nonMasqueradeCIDR: 100.64.0.0/10\n  sshAccess:\n  - 0.0.0.0/0\n  subnets:\n  - cidr: 172.17.17.0/24\n    name: us-east-2a\n    type: Public\n    zone: us-east-2a\n  - cidr: 172.17.18.0/24\n    name: us-east-2b\n    type: Public\n    zone: us-east-2b\n  topology:\n    dns:\n      type: Public\n    masters: public\n    nodes: public\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: master-us-east-2a\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 1\n  minSize: 1\n  nodeLabels:\n    kops.k8s.io/instancegroup: master-us-east-2a\n  role: Master\n  subnets:\n  - us-east-2a\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: nodes\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 6\n  minSize: 6\n  nodeLabels:\n    kops.k8s.io/instancegroup: nodes\n  role: Node\n  subnets:\n  - us-east-2a\n  - us-east-2b\n---\napiVersion: kops/v1alpha2\nkind: SSHCredential\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\nspec:\n  publicKey: \"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\"\n",
						"name": "scoleman"
					}
				}`),
			},
		},
	}
	// Mock Admission Request for a updating a Cluster resource by changing Spec.Name
	// Expect to be rejected by the validating webhook
	AdmissionRequestUpdateDiffName = v1beta1.AdmissionReview{
		TypeMeta: v1.TypeMeta{
			Kind: "AdmissionReview",
		},
		Request: &v1beta1.AdmissionRequest{
			UID: "e911857d-c318-11e8-bbad-025000000001",
			Kind: v1.GroupVersionKind{
				Group:   "cluster-operator.infobloxopen.github.com",
				Version: "v1alpha1",
				Kind:    "Cluster",
			},
			Operation: "UPDATE",
			Object: runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "cluster-operator.infobloxopen.github.com/v1alpha1",
					"kind": "Cluster",
					"metadata": {
						"annotations": {
							"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"cluster-operator.infobloxopen.github.com/v1alpha1\",\"kind\":\"Cluster\",\"metadata\":{\"annotations\":{},\"name\":\"example-cluster\",\"namespace\":\"scoleman\"},\"spec\":{\"config\":\"apiVersion: kops.k8s.io/v1alpha2\\nkind: Cluster\\nmetadata:\\n  name: scoleman.soheil.belamaric.com\\nspec:\\n  api:\\n    dns: {}\\n  authorization:\\n    rbac: {}\\n  channel: stable\\n  cloudProvider: aws\\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\\n  etcdClusters:\\n  - cpuRequest: 200m\\n    etcdMembers:\\n    - instanceGroup: master-us-east-2a\\n      name: a\\n    memoryRequest: 100Mi\\n    name: main\\n  - cpuRequest: 100m\\n    etcdMembers:\\n    - instanceGroup: master-us-east-2a\\n      name: a\\n    memoryRequest: 100Mi\\n    name: events\\n  iam:\\n    allowContainerRegistry: true\\n    legacy: false\\n  kubelet:\\n    anonymousAuth: false\\n  kubernetesApiAccess:\\n  - 0.0.0.0/0\\n  kubernetesVersion: 1.16.7\\n  masterPublicName: api.scoleman.soheil.belamaric.com\\n  networkCIDR: 172.17.16.0/21\\n  networkID: vpc-0a75b33895655b46a\\n  networking:\\n    kubenet: {}\\n  nonMasqueradeCIDR: 100.64.0.0/10\\n  sshAccess:\\n  - 0.0.0.0/0\\n  subnets:\\n  - cidr: 172.17.17.0/24\\n    name: us-east-2a\\n    type: Public\\n    zone: us-east-2a\\n  - cidr: 172.17.18.0/24\\n    name: us-east-2b\\n    type: Public\\n    zone: us-east-2b\\n  topology:\\n    dns:\\n      type: Public\\n    masters: public\\n    nodes: public\\n---\\napiVersion: kops.k8s.io/v1alpha2\\nkind: InstanceGroup\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\n  name: master-us-east-2a\\nspec:\\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\\n  machineType: t2.micro\\n  maxSize: 1\\n  minSize: 1\\n  nodeLabels:\\n    kops.k8s.io/instancegroup: master-us-east-2a\\n  role: Master\\n  subnets:\\n  - us-east-2a\\n---\\napiVersion: kops.k8s.io/v1alpha2\\nkind: InstanceGroup\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\n  name: nodes\\nspec:\\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\\n  machineType: t2.micro\\n  maxSize: 2\\n  minSize: 2\\n  nodeLabels:\\n    kops.k8s.io/instancegroup: nodes\\n  role: Node\\n  subnets:\\n  - us-east-2a\\n  - us-east-2b\\n---\\napiVersion: kops/v1alpha2\\nkind: SSHCredential\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\nspec:\\n  publicKey: \\\"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\\\"\\n\",\"name\":\"scoleman-3\"}}\n"
						},
						"creationTimestamp": "2020-04-22T22:01:40Z",
						"generation": 3,
						"name": "example-cluster",
						"namespace": "scoleman",
						"resourceVersion": "16932",
						"selfLink": "/apis/cluster-operator.infobloxopen.github.com/v1alpha1/namespaces/scoleman/clusters/example-cluster",
						"uid": "5516629b-b095-4975-a316-3339002855b0"
					},
					"spec": {
						"config": "apiVersion: kops.k8s.io/v1alpha2\nkind: Cluster\nmetadata:\n  name: scoleman.soheil.belamaric.com\nspec:\n  api:\n    dns: {}\n  authorization:\n    rbac: {}\n  channel: stable\n  cloudProvider: aws\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\n  etcdClusters:\n  - cpuRequest: 200m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: main\n  - cpuRequest: 100m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: events\n  iam:\n    allowContainerRegistry: true\n    legacy: false\n  kubelet:\n    anonymousAuth: false\n  kubernetesApiAccess:\n  - 0.0.0.0/0\n  kubernetesVersion: 1.16.7\n  masterPublicName: api.scoleman.soheil.belamaric.com\n  networkCIDR: 172.17.16.0/21\n  networkID: vpc-0a75b33895655b46a\n  networking:\n    kubenet: {}\n  nonMasqueradeCIDR: 100.64.0.0/10\n  sshAccess:\n  - 0.0.0.0/0\n  subnets:\n  - cidr: 172.17.17.0/24\n    name: us-east-2a\n    type: Public\n    zone: us-east-2a\n  - cidr: 172.17.18.0/24\n    name: us-east-2b\n    type: Public\n    zone: us-east-2b\n  topology:\n    dns:\n      type: Public\n    masters: public\n    nodes: public\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: master-us-east-2a\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 1\n  minSize: 1\n  nodeLabels:\n    kops.k8s.io/instancegroup: master-us-east-2a\n  role: Master\n  subnets:\n  - us-east-2a\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: nodes\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 2\n  minSize: 2\n  nodeLabels:\n    kops.k8s.io/instancegroup: nodes\n  role: Node\n  subnets:\n  - us-east-2a\n  - us-east-2b\n---\napiVersion: kops/v1alpha2\nkind: SSHCredential\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\nspec:\n  publicKey: \"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\"\n",
						"name": "scoleman"
					}
				}`),
			},
			OldObject: runtime.RawExtension{
				Raw: []byte(`{
					"apiVersion": "cluster-operator.infobloxopen.github.com/v1alpha1",
					"kind": "Cluster",
					"metadata": {
						"annotations": {
							"kubectl.kubernetes.io/last-applied-configuration": "{\"apiVersion\":\"cluster-operator.infobloxopen.github.com/v1alpha1\",\"kind\":\"Cluster\",\"metadata\":{\"annotations\":{},\"name\":\"example-cluster\",\"namespace\":\"scoleman\"},\"spec\":{\"config\":\"apiVersion: kops.k8s.io/v1alpha2\\nkind: Cluster\\nmetadata:\\n  name: scoleman.soheil.belamaric.com\\nspec:\\n  api:\\n    dns: {}\\n  authorization:\\n    rbac: {}\\n  channel: stable\\n  cloudProvider: aws\\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\\n  etcdClusters:\\n  - cpuRequest: 200m\\n    etcdMembers:\\n    - instanceGroup: master-us-east-2a\\n      name: a\\n    memoryRequest: 100Mi\\n    name: main\\n  - cpuRequest: 100m\\n    etcdMembers:\\n    - instanceGroup: master-us-east-2a\\n      name: a\\n    memoryRequest: 100Mi\\n    name: events\\n  iam:\\n    allowContainerRegistry: true\\n    legacy: false\\n  kubelet:\\n    anonymousAuth: false\\n  kubernetesApiAccess:\\n  - 0.0.0.0/0\\n  kubernetesVersion: 1.16.7\\n  masterPublicName: api.scoleman.soheil.belamaric.com\\n  networkCIDR: 172.17.16.0/21\\n  networkID: vpc-0a75b33895655b46a\\n  networking:\\n    kubenet: {}\\n  nonMasqueradeCIDR: 100.64.0.0/10\\n  sshAccess:\\n  - 0.0.0.0/0\\n  subnets:\\n  - cidr: 172.17.17.0/24\\n    name: us-east-2a\\n    type: Public\\n    zone: us-east-2a\\n  - cidr: 172.17.18.0/24\\n    name: us-east-2b\\n    type: Public\\n    zone: us-east-2b\\n  topology:\\n    dns:\\n      type: Public\\n    masters: public\\n    nodes: public\\n---\\napiVersion: kops.k8s.io/v1alpha2\\nkind: InstanceGroup\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\n  name: master-us-east-2a\\nspec:\\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\\n  machineType: t2.micro\\n  maxSize: 1\\n  minSize: 1\\n  nodeLabels:\\n    kops.k8s.io/instancegroup: master-us-east-2a\\n  role: Master\\n  subnets:\\n  - us-east-2a\\n---\\napiVersion: kops.k8s.io/v1alpha2\\nkind: InstanceGroup\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\n  name: nodes\\nspec:\\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\\n  machineType: t2.micro\\n  maxSize: 2\\n  minSize: 2\\n  nodeLabels:\\n    kops.k8s.io/instancegroup: nodes\\n  role: Node\\n  subnets:\\n  - us-east-2a\\n  - us-east-2b\\n---\\napiVersion: kops/v1alpha2\\nkind: SSHCredential\\nmetadata:\\n  labels:\\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\\nspec:\\n  publicKey: \\\"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\\\"\\n\",\"name\":\"scoleman-3\"}}\n"
						},
						"creationTimestamp": "2020-04-22T22:01:40Z",
						"generation": 3,
						"name": "example-cluster",
						"namespace": "scoleman",
						"resourceVersion": "16932",
						"selfLink": "/apis/cluster-operator.infobloxopen.github.com/v1alpha1/namespaces/scoleman/clusters/example-cluster",
						"uid": "5516629b-b095-4975-a316-3339002855b0"
					},
					"spec": {
						"config": "apiVersion: kops.k8s.io/v1alpha2\nkind: Cluster\nmetadata:\n  name: scoleman.soheil.belamaric.com\nspec:\n  api:\n    dns: {}\n  authorization:\n    rbac: {}\n  channel: stable\n  cloudProvider: aws\n  configBase: s3://kops.state.seizadi.infoblox.com/scoleman.soheil.belamaric.com\n  etcdClusters:\n  - cpuRequest: 200m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: main\n  - cpuRequest: 100m\n    etcdMembers:\n    - instanceGroup: master-us-east-2a\n      name: a\n    memoryRequest: 100Mi\n    name: events\n  iam:\n    allowContainerRegistry: true\n    legacy: false\n  kubelet:\n    anonymousAuth: false\n  kubernetesApiAccess:\n  - 0.0.0.0/0\n  kubernetesVersion: 1.16.7\n  masterPublicName: api.scoleman.soheil.belamaric.com\n  networkCIDR: 172.17.16.0/21\n  networkID: vpc-0a75b33895655b46a\n  networking:\n    kubenet: {}\n  nonMasqueradeCIDR: 100.64.0.0/10\n  sshAccess:\n  - 0.0.0.0/0\n  subnets:\n  - cidr: 172.17.17.0/24\n    name: us-east-2a\n    type: Public\n    zone: us-east-2a\n  - cidr: 172.17.18.0/24\n    name: us-east-2b\n    type: Public\n    zone: us-east-2b\n  topology:\n    dns:\n      type: Public\n    masters: public\n    nodes: public\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: master-us-east-2a\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 1\n  minSize: 1\n  nodeLabels:\n    kops.k8s.io/instancegroup: master-us-east-2a\n  role: Master\n  subnets:\n  - us-east-2a\n---\napiVersion: kops.k8s.io/v1alpha2\nkind: InstanceGroup\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\n  name: nodes\nspec:\n  image: kope.io/k8s-1.16-debian-stretch-amd64-hvm-ebs-2020-01-17\n  machineType: t2.micro\n  maxSize: 2\n  minSize: 2\n  nodeLabels:\n    kops.k8s.io/instancegroup: nodes\n  role: Node\n  subnets:\n  - us-east-2a\n  - us-east-2b\n---\napiVersion: kops/v1alpha2\nkind: SSHCredential\nmetadata:\n  labels:\n    kops.k8s.io/cluster: scoleman.soheil.belamaric.com\nspec:\n  publicKey: \"ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD4AK+MI5AqR9lUG+yTlV6l+GrplDcsgOO8gluNdYDB+qGVgjPmZee0jwjYOPdhJPpacDEmSg3ZaCHZOcHJZh5WFpCpri2Ohp2OPg4GtcQI9E6/yRmyPZlLQEC/t93c5bMO29Wv2JuqDJmr09sZErQ1Y4KxxG9VsSK2eoUv9MytTuxh7JdYTTyxvOWmv0/P7WUnmwSaqkgM2QdcGJOb0ulSuWKrh6XS0ShSkG6W2noYEFyl2SCwt4+h4DwcE3lIhkzmMmsWgpFPmRHawLyIH53WAotOJ5pR1gdwz+gp3KAPsldjwRKOkYg0iW9UHxGLUdpIMrCNjvl7kUELSdVnbbGl seizadi@sc-l-seizadi.inca.infoblox.com\"\n",
						"name": "scoleman-updated"
					}
				}`),
			},
		},
	}
)

// Helper function to decode admission review response sent from the validating webhook sever
func decodeResponse(body io.ReadCloser) *v1beta1.AdmissionReview {
	response, _ := ioutil.ReadAll(body)
	review := &v1beta1.AdmissionReview{}
	codecs.UniversalDeserializer().Decode(response, nil, review)
	return review
}

// Helper function to encode an admission review to send to the validating webhook server
func encodeRequest(review *v1beta1.AdmissionReview) []byte {
	ret, err := json.Marshal(review)
	if err != nil {
		fmt.Println(err)
	}
	return ret
}

// Helper function to submit a mock admission review to the validating webhook server
// mock - the mock admission review to send to the server to test
func GetAdmissionReviewForTest(mock v1beta1.AdmissionReview) (*v1beta1.AdmissionReview, error) {
	nsc := &ClusterAdmission{}
	server := httptest.NewServer(GetAdmissionServerNoSSL(nsc, ":8080").Handler)
	requestString := string(encodeRequest(&mock))
	myr := strings.NewReader(requestString)
	r, err := http.Post(server.URL, "application/json", myr)
	if err != nil {
		return nil, err
	}
	review := decodeResponse(r.Body)
	return review, nil
}

// Tests response to insure the validating webook returns a response for the correct object
// Expect UID to match between request and response
func TestServeReturnsCorrectJson(t *testing.T) {
	review, err := GetAdmissionReviewForTest(AdmissionRequestCreate)

	if err != nil {
		t.Error(err)
	}

	if review.Request.UID != AdmissionRequestCreate.Request.UID {
		t.Error("Request and response UID don't match")
	}
}

// Test create case for new Cluster resource
// Expect all create cases to be allowed
func TestCreatePassesValidation(t *testing.T) {
	review, err := GetAdmissionReviewForTest(AdmissionRequestCreate)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(review.Response.Allowed)
	if !review.Response.Allowed {
		t.Error("Create blocked CR, should allow")
	}
}

// Test update Cluster without changing Spec.Name
// Expect update to be allowed
func TestUpdateSameName(t *testing.T) {
	review, err := GetAdmissionReviewForTest(AdmissionRequestUpdateSameName)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(review.Response.Allowed)
	if !review.Response.Allowed {
		t.Error("Update blocked CR with no name change, should allow")
	}
}

// Test update Cluster by changing Spec.Name
// Expect the update to be rejected
func TestUpdateDiffName(t *testing.T) {
	review, err := GetAdmissionReviewForTest(AdmissionRequestUpdateDiffName)

	if err != nil {
		t.Error(err)
	}

	fmt.Println(review.Response.Allowed)
	if review.Response.Allowed {
		t.Error("Update allowed CR with name change, should block")
	}
}
