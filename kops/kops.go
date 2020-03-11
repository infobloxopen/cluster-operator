package kops

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	clusteroperatorv1alpha1 "github.com/seizadi/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"github.com/seizadi/cluster-operator/utils"
	"gopkg.in/yaml.v2"
)

func CreateCluster(cluster clusteroperatorv1alpha1.KopsConfig) (string, error) {

	kopsCmd := "/usr/local/bin/" +
		"kops create cluster" +
		" --name=" + cluster.Name +
		" --state=" + cluster.StateStore +
		// FIXME - Should have ssh-key-name
		" --ssh-public-key=kops.pub" +
		" --vpc=" + cluster.Vpc +
		" --master-count=" + strconv.Itoa(cluster.MasterCount) +
		" --master-size=" + cluster.MasterEc2 +
		" --node-count=" + strconv.Itoa(cluster.WorkerCount) +
		" --node-size=" + cluster.WorkerEc2 +
		" --zones=" + strings.Join(cluster.Zones, ",") +
		" --yes"

	out, err := utils.RunCmd(kopsCmd)
	if err != nil {
		return string(out.Bytes()), err
	}

	return string(out.Bytes()), nil
}

func UpdateCluster(cluster clusteroperatorv1alpha1.KopsConfig) (string, error) {

	kopsCmd := "/usr/local/bin/" +
		"kops update cluster " +
		" --name=" + cluster.Name +
		" --state=" + cluster.StateStore +
		" --yes"

	out, err := utils.RunCmd(kopsCmd)
	if err != nil {
		return string(out.Bytes()), err
	}

	return string(out.Bytes()), nil
}

func DeleteCluster(cluster clusteroperatorv1alpha1.KopsConfig) (string, error) {

	kopsCmd := "/usr/local/bin/" +
		"kops delete cluster --name=" + cluster.Name +
		" --state=" + cluster.StateStore

	out, err := utils.RunCmd(kopsCmd)
	if err != nil {
		return string(out.Bytes()), err
	}

	return string(out.Bytes()), nil
}

func ValidateCluster(cluster clusteroperatorv1alpha1.KopsConfig) (clusteroperatorv1alpha1.KopsStatus, error) {

	status := clusteroperatorv1alpha1.KopsStatus{}

	kopsCmd := "/usr/local/bin/" +
		"kops validate cluster" +
		" --state=" + cluster.StateStore +
		" --name=" + cluster.Name + " -o json"
	out, err := utils.RunCmd(kopsCmd)
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

func GetKubeConfig(cluster clusteroperatorv1alpha1.KopsConfig) (clusteroperatorv1alpha1.KubeConfig, error) {
	kopsCmd := "/usr/local/bin/" +
		"kops export kubecfg --name=" + cluster.Name +
		" --state=" + cluster.StateStore +
		" --kubeconfig=tmp/config.yaml"

	_, err := utils.RunCmd(kopsCmd)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	config := clusteroperatorv1alpha1.KubeConfig{}

	file, err := ioutil.ReadFile("tmp/config.yaml")
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	err = yaml.Unmarshal([]byte(file), &config)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	return config, nil
}
