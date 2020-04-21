package kops

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"github.com/infobloxopen/cluster-operator/utils"
	"gopkg.in/yaml.v2"
)

type KopsCmd struct {
	devMode   bool
	publicKey string
	envs      [][]string
}

func NewKops() (*KopsCmd, error) {
	reqEnvs := []string{
		"AWS_ACCESS_KEY_ID",
		"AWS_SECRET_ACCESS_KEY",
		"KOPS_STATE_STORE",
	}
	filterEnvs := append([]string{
		"SSH_KEY",
		"CLUSTER_OPERATOR_DEVELOPMENT",
	}, reqEnvs[0:]...)

	// FIXME - Integrate public key function into the envs
	k := KopsCmd{
		publicKey: "kops.pub",
		envs:      utils.GetEnvs(filterEnvs),
	}

	for _, pair := range k.envs {
		if pair[0] == "CLUSTER_OPERATOR_DEVELOPMENT" {
			k.devMode = true
		} else if (pair[0] == "SSH_KEY") && (len(pair[1]) > 0) {
			k.publicKey = pair[1]
		}
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
//func (k *KopsCmd) CreateCluster(ctx context.Context, cluster clusteroperatorv1alpha1.KopsConfig) (*utils.Cmd, error) {
//
//	pwd, err := os.Getwd()
//	if err != nil {
//		return nil, err
//	}
//	kopsCmd := "./.bin/docker"
//	kopsArgs := []string {"run"}
//	kopsArgs = append(kopsArgs, utils.GetDockerEnvFlagss(k.envs)...)
//	kopsArgs = append(kopsArgs,
//		"-v " + pwd + "/ssh:/ssh",
//		"soheileizadi/kops:v1.0",
//		"--state=" + cluster.StateStore,
//		"create cluster",
//		"--name=" + cluster.Name,
//		// FIXME - Should have ssh-key-name
//		"--ssh-public-key=" + "/ssh/" + k.publicKey,
//		"--vpc=" + cluster.Vpc,
//		"--master-count=" + strconv.Itoa(cluster.MasterCount),
//		"--master-size=" + cluster.MasterEc2,
//		"--node-count=" + strconv.Itoa(cluster.WorkerCount),
//		"--node-size=" + cluster.WorkerEc2,
//		"--zones=" + strings.Join(cluster.Zones, ","))
//
//	fmt.Println("KOPS CMD ARGS--->>>>>>>>>")
//	fmt.Println(kopsArgs)
//	return utils.New(ctx, nil, kopsCmd, kopsArgs...), nil
//}

// No longer needed when using Kops Manifests
// The flow when using Kops Manifests is always kops replace -> kops update
// The flow applies for both new and existing clusters
//func (k *KopsCmd) CreateCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {
//
//	pwd, err := os.Getwd()
//	if err != nil {
//		return err
//	}
//	kopsCmdStr := "./.bin/" +
//		"docker run" +
//		" -v " + pwd + "/ssh:/ssh " +
//		utils.GetDockerEnvFlags(k.envs) +
//		" soheileizadi/kops:v1.0" +
//		" --state=" + cluster.StateStore +
//		" create cluster" +
//		" --name=" + cluster.Name +
//		// FIXME - Should have ssh-key-name
//		" --ssh-public-key=" + "/ssh/" + k.publicKey +
//		" --vpc=" + cluster.Vpc +
//		" --master-count=" + strconv.Itoa(cluster.MasterCount) +
//		" --master-size=" + cluster.MasterEc2 +
//		" --node-count=" + strconv.Itoa(cluster.WorkerCount) +
//		" --node-size=" + cluster.WorkerEc2 +
//		" --zones=" + strings.Join(cluster.Zones, ",")
//	err = utils.RunStreamingCmd(kopsCmdStr)
//	if err != nil {
//		return err
//	}

//	return nil
//}

func (k *KopsCmd) ReplaceCluster(cluster clusteroperatorv1alpha1.ClusterSpec) error {
	tempConfigFile := cluster.Name + ".yaml"
	err := utils.CopyBufferContentsToTempFile([]byte(cluster.Config), tempConfigFile)
	if err != nil {
		return err
	}

	kopsCmdStr := "./.bin/" +
		"kops replace cluster" +
		" -f tmp/" + tempConfigFile +
		" --state=" + cluster.KopsConfig.StateStore +
		" --force"

	err = utils.RunStreamingCmd(kopsCmdStr)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) UpdateCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {
	if k.devMode { // Dry-run in Dev Mode and skip Update Cluster
		return nil
	}

	kopsCmd := "./.bin/" +
		"kops update cluster " +
		" --state=" + cluster.StateStore +
		" --name=" + cluster.Name +
		// FIXME - Add in when we switch to kops config
		// https://github.com/kubernetes/kops/blob/master/docs/iam_roles.md#use-existing-aws-instance-profiles
		" --lifecycle-overrides IAMRole=ExistsAndWarnIfChanges," +
		"IAMRolePolicy=ExistsAndWarnIfChanges,IAMInstanceProfileRole=ExistsAndWarnIfChanges" +
		" --yes"

	err := utils.RunStreamingCmd(kopsCmd)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) GetCluster(cluster clusteroperatorv1alpha1.KopsConfig) (bool, error) {
	kopsCmd := "./.bin/" +
		"kops get cluster " +
		" --state=" + cluster.StateStore +
		" --name=" + cluster.Name
	exists := true
	err := utils.RunStreamingCmd(kopsCmd)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			exists = false
		}
		return exists, err
	}
	return exists, nil
}

func (k *KopsCmd) RollingUpdateCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {

	// if k.devMode { // Dry-run in Dev Mode and skip Update Cluster
	// 	return nil
	// }

	// Make sure we have config in tmp/config.yaml
	_, err := k.GetKubeConfig(cluster)
	if err != nil {
		return err
	}

	kopsCmd := "./.bin/" +
		"kops rolling-update cluster " +
		" --state=" + cluster.StateStore +
		" --name=" + cluster.Name +
		" --fail-on-validate-error=false" +
		// FIXME - Add in when we switch to kops config
		// https://github.com/kubernetes/kops/blob/master/docs/iam_roles.md#use-existing-aws-instance-profiles
		// " --lifecycle-overrides IAMRole=ExistsAndWarnIfChanges," +
		// "IAMRolePolicy=ExistsAndWarnIfChanges,IAMInstanceProfileRole=ExistsAndWarnIfChanges" +
		" --yes"

	err = utils.RunStreamingCmd(kopsCmd)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) DeleteCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {

	kopsCmd := "./.bin/" +
		"kops delete cluster --name=" + cluster.Name +
		" --state=" + cluster.StateStore +
		" --yes"

	//out, err := utils.RunCmd(kopsCmd)
	err := utils.RunStreamingCmd(kopsCmd)
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

	kopsCmd := "./.bin/" +
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

	if k.devMode { // Dry-run in Dev Mode and skip get kube.config
		return clusteroperatorv1alpha1.KubeConfig{}, nil
	}

	config := clusteroperatorv1alpha1.KubeConfig{}

	kopsCmd := "./.bin/" +
		"kops export kubecfg" +
		" --name=" + cluster.Name +
		" --state=" + cluster.StateStore +
		" --kubeconfig=/tmp/config-" + cluster.Name

	err := utils.RunStreamingCmd(kopsCmd)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	file, err := ioutil.ReadFile("tmp/config-" + cluster.Name)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	err = yaml.Unmarshal([]byte(file), &config)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	return config, nil
}
