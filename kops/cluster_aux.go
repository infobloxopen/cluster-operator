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
	"fmt"
	"io"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/util/pkg/tables"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	getLong = templates.LongDesc(i18n.T(`
	Display one or many resources.` + validResources))

	getExample = templates.Examples(i18n.T(`
	# Get all clusters in a state store
	kops get clusters

	# Get a cluster and its instancegroups
	kops get k8s-cluster.example.com

	# Get a cluster and its instancegroups' YAML desired configuration
	kops get k8s-cluster.example.com -o yaml

	# Save a cluster and its instancegroups' desired configuration to YAML file
	kops get k8s-cluster.example.com -o yaml > cluster-desired-config.yaml

	# Get a secret
	kops get secrets kube -oplaintext

	# Get the admin password for a cluster
	kops get secrets admin -oplaintext`))

	getShort = i18n.T(`Get one or many resources.`)
)

type GetOptions struct {
	output      string
	clusterName string
}

const (
	OutputYaml  = "yaml"
	OutputTable = "table"
	OutputJSON  = "json"
)

// filterClustersByName returns the clusters matching the specified names.
// If names are specified and no cluster is found with a name, we return an error.
func filterClustersByName(clusterNames []string, clusters []*api.Cluster) ([]*api.Cluster, error) {
	if len(clusterNames) != 0 {
		// Build a map as we want to return them in the same order as args
		m := make(map[string]*api.Cluster)
		for _, c := range clusters {
			m[c.ObjectMeta.Name] = c
		}
		var filtered []*api.Cluster
		for _, clusterName := range clusterNames {
			c := m[clusterName]
			if c == nil {
				return nil, fmt.Errorf("cluster not found %q", clusterName)
			}

			filtered = append(filtered, c)
		}
		return filtered, nil
	}

	return clusters, nil
}

func clusterOutputTable(clusters []*api.Cluster, out io.Writer) error {
	t := &tables.Table{}
	t.AddColumn("NAME", func(c *api.Cluster) string {
		return c.ObjectMeta.Name
	})
	t.AddColumn("CLOUD", func(c *api.Cluster) string {
		return c.Spec.CloudProvider
	})
	t.AddColumn("ZONES", func(c *api.Cluster) string {
		zones := sets.NewString()
		for _, s := range c.Spec.Subnets {
			if s.Zone != "" {
				zones.Insert(s.Zone)
			}
		}
		return strings.Join(zones.List(), ",")
	})

	return t.Render(clusters, out, "NAME", "CLOUD", "ZONES")
}

// fullOutputJson outputs the marshalled JSON of a list of clusters and instance groups.  It will handle
// nils for clusters and instanceGroups slices.
func fullOutputJSON(out io.Writer, args ...runtime.Object) error {
	argsLen := len(args)

	if argsLen > 1 {
		if _, err := fmt.Fprint(out, "["); err != nil {
			return err
		}
	}

	for i, arg := range args {
		if i != 0 {
			if _, err := fmt.Fprint(out, ","); err != nil {
				return err
			}
		}
		if err := marshalToWriter(arg, marshalJSON, out); err != nil {
			return err
		}
	}

	if argsLen > 1 {
		if _, err := fmt.Fprint(out, "]"); err != nil {
			return err
		}
	}

	return nil
}

// fullOutputJson outputs the marshalled JSON of a list of clusters and instance groups.  It will handle
// nils for clusters and instanceGroups slices.
func fullOutputYAML(out io.Writer, args ...runtime.Object) error {
	for i, obj := range args {
		if i != 0 {
			if err := writeYAMLSep(out); err != nil {
				return fmt.Errorf("error writing to stdout: %v", err)
			}
		}
		if err := marshalToWriter(obj, marshalYaml, out); err != nil {
			return err
		}
	}
	return nil
}

func fullClusterSpecs(clusters []*api.Cluster) ([]*api.Cluster, error) {
	var fullSpecs []*api.Cluster
	for _, cluster := range clusters {
		configBase, err := registry.ConfigBase(cluster)
		if err != nil {
			return nil, fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
		}
		fullSpec := &api.Cluster{}
		err = registry.ReadConfigDeprecated(configBase.Join(registry.PathClusterCompleted), fullSpec)
		if err != nil {
			return nil, fmt.Errorf("error reading full cluster spec for %q: %v", cluster.ObjectMeta.Name, err)
		}
		fullSpecs = append(fullSpecs, fullSpec)
	}
	return fullSpecs, nil
}

func writeYAMLSep(out io.Writer) error {
	_, err := out.Write([]byte("\n---\n\n"))
	if err != nil {
		return fmt.Errorf("error writing to stdout: %v", err)
	}
	return nil
}

type marshalFunc func(obj runtime.Object) ([]byte, error)

func marshalToWriter(obj runtime.Object, marshal marshalFunc, w io.Writer) error {
	b, err := marshal(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return fmt.Errorf("error writing to stdout: %v", err)
	}
	return nil
}

// obj must be a pointer to a marshalable object
func marshalYaml(obj runtime.Object) ([]byte, error) {
	y, err := kopscodecs.ToVersionedYaml(obj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling yaml: %v", err)
	}
	return y, nil
}

// obj must be a pointer to a marshalable object
func marshalJSON(obj runtime.Object) ([]byte, error) {
	j, err := kopscodecs.ToVersionedJSON(obj)
	if err != nil {
		return nil, fmt.Errorf("error marshaling json: %v", err)
	}
	return j, nil
}

func int32PointerToString(v *int32) string {
	if v == nil {
		return "-"
	}
	return strconv.Itoa(int(*v))
}
