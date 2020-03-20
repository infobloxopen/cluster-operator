package cluster

import (
	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"testing"
)

type testZone struct {
	value string
	found bool
}

func CheckKopsDefaultConfigTest(t *testing.T) {
	instance := &clusteroperatorv1alpha1.Cluster{}
	instance.Name = "TEST"
	
	defaultConfig := clusteroperatorv1alpha1.KopsConfig{
		// FIXME - Pickup DNS zone from Operator Config
		Name:        instance.Name + ".soheil.belamaric.com",
		MasterCount: 1,
		MasterEc2:   "t2.micro",
		WorkerCount: 2,
		WorkerEc2:   "t2.micro",
		// FIXME - Pickup state store from Operator Config
		StateStore:  "s3://kops.state.seizadi.infoblox.com",
		Vpc:         "vpc-0a75b33895655b46a",
		Zones:       []string{"us-east-2a", "us-east-2b"},
	}
	
	testZones := []testZone {}
	for _, z := range defaultConfig.Zones {
		testZones = append(testZones, testZone{ z, false})
	}
	
	config := CheckKopsDefaultConfig(instance.Spec)
	
	if (config.Name != defaultConfig.Name) {
		t.Error("Expected ", defaultConfig.Name, "got ", config.Name)
	}
	
	if (config.MasterCount != defaultConfig.MasterCount) {
		t.Error("Expected ", defaultConfig.MasterCount, "got ", config.MasterCount)
	}
	
	if (config.MasterEc2 != defaultConfig.MasterEc2) {
		t.Error("Expected ", defaultConfig.MasterEc2, "got ", config.MasterEc2)
	}
	
	if (config.WorkerCount != defaultConfig.WorkerCount) {
		t.Error("Expected ", defaultConfig.WorkerCount, "got ", config.WorkerCount)
	}
	
	if (config.WorkerEc2 != defaultConfig.WorkerEc2) {
		t.Error("Expected ", defaultConfig.WorkerEc2, "got ", config.WorkerEc2)
	}
	
	if (config.StateStore != defaultConfig.StateStore) {
		t.Error("Expected ", defaultConfig.StateStore, "got ", config.StateStore)
	}
	
	if (config.Vpc != defaultConfig.Vpc) {
		t.Error("Expected ", defaultConfig.Vpc, "got ", config.Vpc)
	}
	
	for _, z := range testZones {
		if ( z.found == false ) {
			t.Error("Zone ", z.value, " not found")
		}
	}
	
	if (len(config.Zones) != len(defaultConfig.Zones)) {
		t.Error("Expected Zone size", len(defaultConfig.Zones), "got ", len(config.Zones))
	}
}
