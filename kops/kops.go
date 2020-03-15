package kops

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"errors"

	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"github.com/infobloxopen/cluster-operator/utils"
	"gopkg.in/yaml.v2"
)

type KopsCmd struct {
	publicKey string
	envs      [][]string
}

func NewKops() (*KopsCmd, error) {
	reqEnvs := []string {
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"KOPS_STATE_STORE",
	}
	filterEnvs := append([]string {
		"SSH_KEY",
	}, reqEnvs[0:]...)

	// FIXME - Integrate public key function into the envs
	k := KopsCmd {
		publicKey: "kops.pub",
		envs: utils.GetEnvs(filterEnvs),
	}
	envKey := os.Getenv("SSH_KEY")
	if len(envKey) != 0 {
		k.publicKey = envKey
	}
	
	missingEnvs := utils.CheckEnvs(k.envs, reqEnvs)
	
	if len(missingEnvs) > 0 {
		foundEnvs := []string{}
		for _, e := range k.envs {
			foundEnvs = append(foundEnvs, e[0])
		}
		return &k, errors.New("Missing environment variables for Kops " + strings.Join(missingEnvs, ", ") +
			" Found Envs" + strings.Join(foundEnvs, ", "))
	}
	
	return &k, nil
}

// CreateCluster provisions a new cluster
func (k *KopsCmd) CreateCluster(ctx context.Context, cluster clusteroperatorv1alpha1.KopsConfig) (*utils.Cmd, error) {

	kopsCmdStr := "/usr/local/bin/" +
		"kops create cluster" +
		" --name=" + cluster.Name +
		" --state=" + cluster.StateStore +
		// FIXME - Should have ssh-key-name
		" --ssh-public-key=" + k.publicKey +
		" --vpc=" + cluster.Vpc +
		" --master-count=" + strconv.Itoa(cluster.MasterCount) +
		" --master-size=" + cluster.MasterEc2 +
		" --node-count=" + strconv.Itoa(cluster.WorkerCount) +
		" --node-size=" + cluster.WorkerEc2 +
		" --zones=" + strings.Join(cluster.Zones, ",")

	kopsCmd := strings.Split(kopsCmdStr, " ")
	return utils.New(ctx, nil, kopsCmd[0], kopsCmd[1:]...), nil
}

func (k *KopsCmd) UpdateCluster(cluster clusteroperatorv1alpha1.KopsConfig) (string, error) {

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

func (k *KopsCmd) DeleteCluster(cluster clusteroperatorv1alpha1.KopsConfig) (string, error) {

	kopsCmd := "/usr/local/bin/" +
		"kops delete cluster --name=" + cluster.Name +
		" --state=" + cluster.StateStore +
		" --yes"

	out, err := utils.RunCmd(kopsCmd)
	if err != nil {
		return string(out.Bytes()), err
	}

	return string(out.Bytes()), nil
}

func (k *KopsCmd) ValidateCluster(cluster clusteroperatorv1alpha1.KopsConfig) (clusteroperatorv1alpha1.KopsStatus, error) {

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

func (k *KopsCmd) GetKubeConfig(cluster clusteroperatorv1alpha1.KopsConfig) (clusteroperatorv1alpha1.KubeConfig, error) {
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
