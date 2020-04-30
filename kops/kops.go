package kops

import (
	"bytes"
	"encoding/json"
	"fmt"
	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"github.com/infobloxopen/cluster-operator/utils"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

type KopsCmd struct {
	devMode         bool
	publicKey       string
	runStreamingCmd func(string) error
	runCmd          func(string) (*bytes.Buffer, error)
}

func NewKops() (*KopsCmd, error) {
	k := KopsCmd{
		publicKey:       viper.GetString("kops.ssh.key"),
		devMode:         viper.GetBool("development"),
		runStreamingCmd: utils.RunStreamingCmd,
		runCmd:          utils.RunCmd,
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
//	kopsCmd := "/usr/local/bin/docker"
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
//	kopsCmdStr := "/usr/local/bin/" +
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

	kopsCmd := viper.GetString("docker.bin.path") + " run" +
		// FIXME Don't think we will need this anymore
		// " -v " + pwd + "/ssh:/ssh " +
		" -v " + viper.GetString("tmp.dir") + "/" + tempConfigFile + ":" + viper.GetString("kops.kube.dir") + "/" + tempConfigFile +
		utils.GetDockerEnvFlags() +
		viper.GetString("kops.container") +
		" replace cluster" +
		" -f " + viper.GetString("kops.kube.dir") + "/" + tempConfigFile +
		// FIXME Do we want to use state.store or cluster.StateStore
		" --state=" + viper.GetString("kops.state.store") +
		" --force"

	err = k.runStreamingCmd(kopsCmd)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) UpdateCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {
	if k.devMode { // Dry-run in Dev Mode and skip Update Cluster
		return nil
	}

	kopsCmd := viper.GetString("docker.bin.path") + " run" +
		utils.GetDockerEnvFlags() +
		viper.GetString("kops.container") +
		" update cluster " +
		" --state=" + viper.GetString("kops.state.store") +
		" --name=" + cluster.Name +
		// FIXME - Add in when we switch to kops config
		// https://github.com/kubernetes/kops/blob/master/docs/iam_roles.md#use-existing-aws-instance-profiles
		" --lifecycle-overrides IAMRole=ExistsAndWarnIfChanges," +
		" IAMRolePolicy=ExistsAndWarnIfChanges,IAMInstanceProfileRole=ExistsAndWarnIfChanges" +
		" --yes"

	err := k.runStreamingCmd(kopsCmd)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) GetCluster(cluster clusteroperatorv1alpha1.KopsConfig) (bool, error) {
	kopsCmd := viper.GetString("docker.bin.path") +
		" run" +
		utils.GetDockerEnvFlags() +
		viper.GetString("kops.container") +
		" get cluster " +
		" --state=" + viper.GetString("kops.state.store") +
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

	if k.devMode { // Dry-run in Dev Mode and skip Update Cluster
		return nil
	}

	// Make sure we have config in tmp/config.yaml
	_, err := k.GetKubeConfig(cluster)
	if err != nil {
		return err
	}

	kopsCmd := viper.GetString("docker.bin.path") +
		" run" + utils.GetDockerEnvFlags() +
		" -e KUBECONFIG=" + viper.GetString("kops.kube.dir") + "/config.yaml" +
		" -v " + viper.GetString("tmp.dir") + ":" + viper.GetString("kops.kube.dir") +
		viper.GetString("kops.container") +
		" rolling-update cluster " +
		" --state=" + viper.GetString("kops.state.store") +
		" --name=" + cluster.Name +
		//FIXME - Add in when we switch to kops config
		// https://github.com/kubernetes/kops/blob/master/docs/iam_roles.md#use-existing-aws-instance-profiles
		// " --lifecycle-overrides IAMRole=ExistsAndWarnIfChanges," +
		// "IAMRolePolicy=ExistsAndWarnIfChanges,IAMInstanceProfileRole=ExistsAndWarnIfChanges" +
		" --yes"

	err = k.runStreamingCmd(kopsCmd)
	if err != nil {
		return err
	}

	return nil
}

func (k *KopsCmd) DeleteCluster(cluster clusteroperatorv1alpha1.KopsConfig) error {

	kopsCmd := viper.GetString("docker.bin.path") + " run" +
		utils.GetDockerEnvFlags() +
		viper.GetString("kops.container") +
		" --state=" + cluster.StateStore +
		" delete cluster --name=" + cluster.Name +
		" --yes"

	//out, err := utils.RunCmd(kopsCmd)
	err := k.runStreamingCmd(kopsCmd)
	if err != nil {
		return err
	}

	return nil
}

//func (k *KopsCmd) DeleteCluster(cluster clusteroperatorv1alpha1.KopsConfig) (string, error) {
//
//	//kopsCmd := "/usr/local/bin/docker"
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

	kopsCmd := viper.GetString("docker.bin.path") + " run" +
		utils.GetDockerEnvFlags() +
		" -e KUBECONFIG=" + viper.GetString("kops.kube.dir") + "/config.yaml" +
		" -v " + viper.GetString("tmp.dir") + ":" + viper.GetString("kops.kube.dir") +
		viper.GetString("kops.container") +
		" validate cluster" +
		" --state=" + cluster.StateStore +
		" --name=" + cluster.Name + " -o json"
	out, err := k.runCmd(kopsCmd)
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

	kopsCmd := viper.GetString("docker.bin.path") + " run" +
		" -v " + viper.GetString("tmp.dir") + ":" + viper.GetString("kops.kube.dir") +
		utils.GetDockerEnvFlags() +
		viper.GetString("kops.container") +
		" --state=" + cluster.StateStore +
		" export kubecfg --name=" + cluster.Name +
		" --kubeconfig=" + viper.GetString("kops.kube.dir") + "/config.yaml"

	err := k.runStreamingCmd(kopsCmd)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	file, err := ioutil.ReadFile(viper.GetString("tmp.dir") + "/config.yaml")
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	err = yaml.Unmarshal([]byte(file), &config)
	if err != nil {
		return clusteroperatorv1alpha1.KubeConfig{}, err
	}

	return config, nil
}
