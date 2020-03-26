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
	"io/ioutil"

	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kops/cmd/kops/util"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kubectl/pkg/util/i18n"
	"k8s.io/kubectl/pkg/util/templates"
)

var (
	exportKubecfgLong = templates.LongDesc(i18n.T(`
	Export a kubecfg file for a cluster from the state store. The configuration
	will be saved into a users $HOME/.kube/config file.
	To export the kubectl configuration to a specific file set the KUBECONFIG
	environment variable.`))

	exportKubecfgExample = templates.Examples(i18n.T(`
	# export a kubecfg file
	kops export kubecfg kubernetes-cluster.example.com
		`))

	exportKubecfgShort = i18n.T(`Export kubecfg.`)
)

type ExportKubecfgOptions struct {
	KubeConfigPath string
	all            bool
}

func RunExportKubecfg(f *util.Factory, out io.Writer, options *ExportKubecfgOptions, args []string) (*clusteroperatorv1alpha1.KubeConfig, error) {
	clientset, err := f.Clientset()
	if err != nil {
		return nil, err
	}

	var clusterList []*api.Cluster
	if options.all {
		if len(args) != 0 {
			return nil, fmt.Errorf("Cannot use both --all flag and positional arguments")
		}
		list, err := clientset.ListClusters(metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		for i := range list.Items {
			clusterList = append(clusterList, &list.Items[i])
		}
	} else {
		err := rootCommand.ProcessArgs(args)
		if err != nil {
			return nil, err
		}
		cluster, err := GetCluster(f, args[0])
		if err != nil {
			return nil, err
		}
		clusterList = append(clusterList, cluster)
	}

	for _, cluster := range clusterList {
		keyStore, err := clientset.KeyStore(cluster)
		if err != nil {
			return nil, err
		}

		secretStore, err := clientset.SecretStore(cluster)
		if err != nil {
			return nil, err
		}

		conf, err := kubeconfig.BuildKubecfg(cluster, keyStore, secretStore, &commands.CloudDiscoveryStatusStore{}, buildPathOptions(options))
		if err != nil {
			return nil, err
		}

		if err := conf.WriteKubecfg(); err != nil {
			return nil, err
		}
	}

	file, err := ioutil.ReadFile("tmp/config-" + args[0])
	if err != nil {
		return nil, err
	}
	config := clusteroperatorv1alpha1.KubeConfig{}
	err = yaml.Unmarshal([]byte(file), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func buildPathOptions(options *ExportKubecfgOptions) *clientcmd.PathOptions {
	pathOptions := clientcmd.NewDefaultPathOptions()

	if len(options.KubeConfigPath) > 0 {
		pathOptions.GlobalFile = options.KubeConfigPath
		pathOptions.EnvVar = ""
		pathOptions.GlobalFileSubpath = ""
	}

	return pathOptions
}
