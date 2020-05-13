package kops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"github.com/infobloxopen/cluster-operator/utils"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

type KopsCmd struct {
	devMode         bool
	publicKey       string
	runStreamingCmd func(string) error
	runCmd          func(string) (*bytes.Buffer, error)
	path string
}

func NewKops() (*KopsCmd, error) {
	k := KopsCmd{
		publicKey:       viper.GetString("kops.ssh.key"),
		devMode:         viper.GetBool("development"),
		runStreamingCmd: utils.RunStreamingCmd,
		runCmd:          utils.RunCmd,
		path: viper.GetString("kops.path"),
	}

	return &k, nil
}

func (k *KopsCmd) ReplaceCluster(cluster clusteroperatorv1alpha1.ClusterSpec) error {
	tempConfigFile := cluster.Name + ".yaml"
	err := utils.CopyBufferContentsToTempFile([]byte(cluster.Config), tempConfigFile)
	if err != nil {
		return err
	}

	kopsCmdStr := k.path +
		" replace cluster" +
		" -f ." + viper.GetString("kops.kube.dir") + "/" + tempConfigFile +
		" --state=" + viper.GetString("kops.state.store") +
		" --force"

	err = k.runStreamingCmd(kopsCmdStr)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) UpdateCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {
	if k.devMode { // Dry-run in Dev Mode and skip Update Cluster
		return nil
	}

	kopsCmdStr := k.path +
		" update cluster " +
		" --state=" + viper.GetString("kops.state.store") +
		" --name=" + cluster.Name +
		// FIXME - Add in when we switch to kops config
		// https://github.com/kubernetes/kops/blob/master/docs/iam_roles.md#use-existing-aws-instance-profiles
		// " --lifecycle-overrides IAMRole=ExistsAndWarnIfChanges," +
		// "IAMRolePolicy=ExistsAndWarnIfChanges,IAMInstanceProfileRole=ExistsAndWarnIfChanges" +
		" --yes"

	err := k.runStreamingCmd(kopsCmdStr)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) GetCluster(cluster clusteroperatorv1alpha1.KopsConfig) (bool, error) {
	kopsCmdStr := k.path +
		" get cluster " +
		" --state=" + viper.GetString("kops.state.store") +
		" --name=" + cluster.Name
	exists := true
	err := k.runStreamingCmd(kopsCmdStr)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			exists = false
		}
		return exists, err
	}
	return exists, nil
}

func (k *KopsCmd) RollingUpdateCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {

	if k.devMode { // Dry-run in Dev Mode and skip Update Cluster
		return nil
	}

	// Make sure we have config in tmp/config.yaml
	_, err := k.GetKubeConfig(cluster)
	if err != nil {
		return err
	}

	kopsCmdStr := k.path +
		" rolling-update cluster " +
		" --state=" + viper.GetString("kops.state.store") +
		" --name=" + cluster.Name +
		" --fail-on-validate-error=false" +
		// FIXME - Add in when we switch to kops config
		// https://github.com/kubernetes/kops/blob/master/docs/iam_roles.md#use-existing-aws-instance-profiles
		// " --lifecycle-overrides IAMRole=ExistsAndWarnIfChanges," +
		// "IAMRolePolicy=ExistsAndWarnIfChanges,IAMInstanceProfileRole=ExistsAndWarnIfChanges" +
		" --yes"

	err = k.runStreamingCmd(kopsCmdStr)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) DeleteCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {

	kopsCmdStr := k.path +
		" delete cluster --name=" + cluster.Name +
		" --state=" + viper.GetString("kops.state.store") +
		" --yes"

	//out, err := utils.RunCmd(kopsCmd)
	err := k.runStreamingCmd(kopsCmdStr)
	if err != nil {
		return err
	}

	return nil
}

//func (k *KopsCmd) DeleteCluster(cluster clusteroperatorv1alpha1.KopsConfig) (string, error) {
//
//	//kopsCmd := "./.bin/docker"
//	kopsArgs := []string{"run", "--env-file=tmp/kops_env"}
//	kopsArgs = append(kopsArgs,
//		"soheileizadi/kops:v1.0",
//		"delete",
//		"cluster",
//		"--name=" + cluster.Name,
//		"--yes")
//
//	fmt.Println(kopsArgs)
//	out, err := utils.RunDockerCmd(kopsArgs)
//	if err != nil {
//		return string(out.Bytes()), err
//	}
//
//	return string(out.Bytes()), nil
//}

func (k *KopsCmd) ValidateCluster(cluster clusteroperatorv1alpha1.KopsConfig) (clusteroperatorv1alpha1.KopsStatus, error) {

	status := clusteroperatorv1alpha1.KopsStatus{}

	if k.devMode { // Dry-run in Dev Mode and skip Validate Cluster return Cluster Up Status
		status = clusteroperatorv1alpha1.KopsStatus{
			Nodes: []clusteroperatorv1alpha1.KopsNode{
				{
					Name:     "ip-172-17-17-143.compute.internal",
					Zone:     "us-east-2a",
					Role:     "Master",
					Hostname: "ip-172-17-17-143.compute.internal",
					Status:   "True",
				},
			},
		}
		return status, nil
	}

	// Make sure we have config in tmp/config.yaml
	_, err := k.GetKubeConfig(cluster)
	if err != nil {
		return status, err
	}

	kopsCmdStr := k.path +
		" validate cluster" +
		" --state=" + viper.GetString("kops.state.store") +
		" --name=" + cluster.Name + " -o json"
	out, err := k.runCmd(kopsCmdStr)
	if err != nil {
		return status, err
	}

	json.Unmarshal(out.Bytes(), &status)
	if err != nil {
		return status, err
	}

	fmt.Println("Kops Response: ", string(out.Bytes()))
	return status, nil
}

func (k *KopsCmd) GetKubeConfig(cluster clusteroperatorv1alpha1.KopsConfig) (clusteroperatorv1alpha1.KubeConfig, error) {

	if k.devMode { // Dry-run in Dev Mode and skip get kube.config
		return clusteroperatorv1alpha1.KubeConfig{}, nil
	}

	config := clusteroperatorv1alpha1.KubeConfig{}

	kopsCmdStr := k.path +
		" export kubecfg" +
		" --name=" + cluster.Name +
		" --state=" + viper.GetString("kops.state.store") +
		" --kubeconfig=" +  viper.GetString("tmp.dir") + "/config-" + cluster.Name

	err := k.runStreamingCmd(kopsCmdStr)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	file, err := ioutil.ReadFile(viper.GetString("tmp.dir") + "/config-" + cluster.Name)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	err = yaml.Unmarshal([]byte(file), &config)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	return config, nil
}

func (k *KopsCmd) ListClusters(stateStore string) ([]string, error) {
	kopsCmdStr := k.path +
		" get cluster " +
		" --state=" + viper.GetString("kops.state.store") +
		" -o json | jq -r '.[][\"metadata\"][\"name\"]'"

	out, err := k.runCmd(kopsCmdStr)
	if err != nil {
		return nil, err
	}

	return strings.Split(string(out.Bytes()), "\n"), nil
}
