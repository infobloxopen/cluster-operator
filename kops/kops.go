package kops

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"os/exec"
	clusteroperatorv1alpha1 "github.com/seizadi/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
)

func CreateCluster(cluster clusteroperatorv1alpha1.KopsConfig) (string, error) {
	
	var out bytes.Buffer
	
	kopsCmd := "/usr/local/bin/" +
		"kops create cluster" +
		" --state=" + cluster.StateStore +
		" --vpc="  + cluster.Vpc +
		" --node-count=" + strconv.Itoa(cluster.WorkerCount) +
		" --master-size=" + cluster.MasterEc2 +
		" --node-size=" + cluster.WorkerEc2 +
		" --zones=" + strings.Join(cluster.Zones, ", ") +
		" --name=" + cluster.Name +
		" --master-count=" + strconv.Itoa(cluster.MasterCount) +
		" --yes"
	cmd := exec.Command(kopsCmd)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	
	return string(out.Bytes()), nil
}

func ValidateCluster(cluster clusteroperatorv1alpha1.KopsConfig) (clusteroperatorv1alpha1.KopsStatus, error) {
	
	var out bytes.Buffer
	status := clusteroperatorv1alpha1.KopsStatus {}
	
	kopsCmd := "/usr/local/bin/" +
		"kops validate cluster" +
		" --state=" + cluster.StateStore +
		" --name=" + cluster.Name
	cmd := exec.Command(kopsCmd)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return status, err
	}
	
	json.Unmarshal(out.Bytes(), &status)
	if err != nil {
		return status, err
	}
	
	return status, nil
}
