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

package kops

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"k8s.io/kops/upup/pkg/fi/cloudup"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/util/pkg/tables"
)

type ValidateClusterOptions struct {
	output     string
	wait       time.Duration
	Kubeconfig string
}

func (o *ValidateClusterOptions) InitDefaults() {
	o.output = OutputTable
}

func RunValidateCluster(f *util.Factory, args []string, out io.Writer, options *ValidateClusterOptions) (*clusteroperatorv1alpha1.KopsStatus, error) {
	err := rootCommand.ProcessArgs(args)
	if err != nil {
		return nil, err
	}

	cluster, err := GetCluster(f, args[0])
	if err != nil {
		return nil, err
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	clientSet, err := f.Clientset()
	if err != nil {
		return nil, err
	}

	list, err := clientSet.InstanceGroupsFor(cluster).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("cannot get InstanceGroups for %q: %v", cluster.ObjectMeta.Name, err)
	}

	if options.output == OutputTable {
		fmt.Fprintf(out, "Validating cluster %v\n\n", cluster.ObjectMeta.Name)
	}

	var instanceGroups []api.InstanceGroup
	for _, ig := range list.Items {
		instanceGroups = append(instanceGroups, ig)
		klog.V(2).Infof("instance group: %#v\n\n", ig.Spec)
	}

	if len(instanceGroups) == 0 {
		return nil, fmt.Errorf("no InstanceGroup objects found")
	}

	// TODO: Refactor into util.Factory
	contextName := cluster.ObjectMeta.Name
	configLoadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if options.Kubeconfig != "" {
		configLoadingRules.ExplicitPath = options.Kubeconfig
	}
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		configLoadingRules,
		&clientcmd.ConfigOverrides{CurrentContext: contextName}).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot load kubecfg settings for %q: %v", contextName, err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot build kubernetes api client for %q: %v", contextName, err)
	}

	validator, err := validation.NewClusterValidator(cluster, cloud, list, k8sClient)
	if err != nil {
		return nil, fmt.Errorf("unexpected error creating validatior: %v", err)
	}

	result, err := validator.Validate()
	if err != nil {
		return nil, fmt.Errorf("unexpected error during validation: %v", err)
	}

	switch options.output {
	case OutputTable:
		if err := validateClusterOutputTable(result, cluster, instanceGroups, out); err != nil {
			return nil, err
		}

	case OutputYaml:
		y, err := yaml.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal YAML: %v", err)
		}
		if _, err := out.Write(y); err != nil {
			return nil, fmt.Errorf("error writing to output: %v", err)
		}

	case OutputJSON:
		j, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal JSON: %v", err)
		}
		if _, err := out.Write(j); err != nil {
			return nil, fmt.Errorf("error writing to output: %v", err)
		}

	default:
		return nil, fmt.Errorf("unknown output format: %q", options.output)
	}

	//copy results into our KopsStatus form
	//cannot use kops struct in cluster_types.go since no deep copy exists
	var status clusteroperatorv1alpha1.KopsStatus

	//if no failures, copy nodes to track in kops status
	if len(result.Failures) == 0 {
		if len(result.Nodes) > 0 {
			for n := 0; n < len(result.Nodes); n++ {
				status.Nodes = append(status.Nodes, clusteroperatorv1alpha1.KopsNode{Name: result.Nodes[n].Name, Zone: result.Nodes[n].Zone,
					Role: result.Nodes[n].Role, Hostname: result.Nodes[n].Hostname, Status: string(result.Nodes[n].Status)})
			}
		}
		return &status, nil
	}

	//if there are failures, return them so we can set kops status
	for f := 0; f < len(result.Failures); f++ {
		status.Failures[f].Kind = result.Failures[f].Kind
		status.Failures[f].Name = result.Failures[f].Name
		status.Failures[f].Message = result.Failures[f].Message
	}
	return &status, nil

}

func validateClusterOutputTable(result *validation.ValidationCluster, cluster *api.Cluster, instanceGroups []api.InstanceGroup, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("NAME", func(c api.InstanceGroup) string {
		return c.ObjectMeta.Name
	})
	t.AddColumn("ROLE", func(c api.InstanceGroup) string {
		return string(c.Spec.Role)
	})
	t.AddColumn("MACHINETYPE", func(c api.InstanceGroup) string {
		return c.Spec.MachineType
	})
	t.AddColumn("SUBNETS", func(c api.InstanceGroup) string {
		return strings.Join(c.Spec.Subnets, ",")
	})
	t.AddColumn("MIN", func(c api.InstanceGroup) string {
		return int32PointerToString(c.Spec.MinSize)
	})
	t.AddColumn("MAX", func(c api.InstanceGroup) string {
		return int32PointerToString(c.Spec.MaxSize)
	})

	fmt.Fprintln(out, "INSTANCE GROUPS")
	err := t.Render(instanceGroups, out, "NAME", "ROLE", "MACHINETYPE", "MIN", "MAX", "SUBNETS")
	if err != nil {
		return fmt.Errorf("cannot render nodes for %q: %v", cluster.Name, err)
	}

	{
		nodeTable := &tables.Table{}
		nodeTable.AddColumn("NAME", func(n *validation.ValidationNode) string {
			return n.Name
		})

		nodeTable.AddColumn("READY", func(n *validation.ValidationNode) v1.ConditionStatus {
			return n.Status
		})

		nodeTable.AddColumn("ROLE", func(n *validation.ValidationNode) string {
			return n.Role
		})

		fmt.Fprintln(out, "\nNODE STATUS")
		if err := nodeTable.Render(result.Nodes, out, "NAME", "ROLE", "READY"); err != nil {
			return fmt.Errorf("cannot render nodes for %q: %v", cluster.Name, err)
		}
	}

	if len(result.Failures) != 0 {
		failuresTable := &tables.Table{}
		failuresTable.AddColumn("KIND", func(e *validation.ValidationError) string {
			return e.Kind
		})
		failuresTable.AddColumn("NAME", func(e *validation.ValidationError) string {
			return e.Name
		})
		failuresTable.AddColumn("MESSAGE", func(e *validation.ValidationError) string {
			return e.Message
		})

		fmt.Fprintln(out, "\nVALIDATION ERRORS")
		if err := failuresTable.Render(result.Failures, out, "KIND", "NAME", "MESSAGE"); err != nil {
			return fmt.Errorf("error rendering failures table: %v", err)
		}

		fmt.Fprintf(out, "\nValidation Failed\n")
	} else {
		fmt.Fprintf(out, "\nYour cluster %s is ready\n", cluster.Name)
	}

	return nil
}
