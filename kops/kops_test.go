package kops

import (
	"bytes"
	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"strings"
	"testing"
)

var kopsConfig clusteroperatorv1alpha1.KopsConfig = clusteroperatorv1alpha1.KopsConfig{
	Name:        "test.soheil.belamaric.com",
	MasterCount: 1,
	MasterEc2:   "t2.micro",
	WorkerCount: 2,
	WorkerEc2:   "t2.micro",
	StateStore:  "s3://kops.state.seizadi.infoblox.com",
	Vpc:         "vpc-0a75b33895655b46a",
	Zones:       []string{"us-east-2a", "us-east-2b"},
}

type testCase struct {
	value string
	found bool
}

var cmd string

func mockRunStreamingCmd(cmdString string) error {
	cmd = cmdString
	return nil
}

func mockRunCmd(cmdString string) (*bytes.Buffer, error) {
	cmd = cmdString
	return nil, nil
}

func TestCreateCluster(t *testing.T) {
	k, err := NewKops()
	if err != nil {
		t.Error("Expected no error got", err)
		return
	}

	k.runStreamingCmd = mockRunStreamingCmd
	k.runCmd = mockRunCmd

	values := []testCase{
		{"replace", false},
		{"cluster", false},
		{"-f", false},
		{"/TestCluster.yaml", false},
		{"--state=", false},
		{"--force", false},
	}

	cluster := clusteroperatorv1alpha1.ClusterSpec{
		Name:       "TestCluster",
		Config:     "",
		KopsConfig: clusteroperatorv1alpha1.KopsConfig{},
	}

	err = k.ReplaceCluster(cluster)
	if err != nil {
		t.Error("Expected no error got", err)
		return
	}

	cmdValues := strings.Split(cmd, " ")
	for _, c := range cmdValues {
		for i, v := range values {
			if v.value == c {
				values[i].found = true
				break
			}
		}
	}

	for _, v := range values {
		if v.found == false {
			t.Error("Expected ", v.value, "not found")
		}
	}
}
