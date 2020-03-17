package kops

import (
	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
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

// FIXME - Put this back in when we figure out what to do about runing Cmd.Run() native
//func TestCreateCluster(t *testing.T)  {
//	k, _ := NewKops()
//	k.envs = [][]string {{"key1", "value1"}, {"key2", "value2"}}
//	values := []testCase{
//		{"/usr/local/bin/docker", false},
//		{"run", false},
//		{ "-e", false},
//		{ "key1=value1", false},
//		{ "key2=value2", false},
//		{"create", false},
//		{"cluster", false},
//		{"--name=" + kopsConfig.Name, false},
//		{"--state=" + kopsConfig.StateStore, false},
//		{"--ssh-public-key=" + k.publicKey, false},
//		{"--vpc=" + kopsConfig.Vpc, false},
//		{"--master-count=" + strconv.Itoa(kopsConfig.MasterCount), false},
//		{"--master-size=" + kopsConfig.MasterEc2, false},
//		{"--node-count=" + strconv.Itoa(kopsConfig.WorkerCount), false},
//		{"--node-size=" + kopsConfig.WorkerEc2, false},
//		{"--zones=" + strings.Join(kopsConfig.Zones, ","), false},
//	}
//
//	cmd, _ := k.CreateCluster(context.TODO(), kopsConfig)
//	cmdString := cmd.GetCmdString()
//
//	for _, c := range cmdString {
//		for i, v := range values {
//			if (v.value == c) {
//				values[i].found = true
//				break
//			}
//		}
//	}
//
//	for _, v := range values {
//		if (v.found == false) {
//			t.Error("Expected ", v.value, "not found")
//		}
//	}
//}
