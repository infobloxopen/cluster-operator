package utils

import (
	"os"
	"testing"
)

func TestGetEnvs(t *testing.T) {
	key := "CLUSTEROP_TEST_9898"
	value := "2020"
	os.Setenv(key, value)
	envs := GetEnvs([]string{key})
	
	if len(envs) != 1 {
		t.Error("Expected 1 got ", len(envs))
	} else {
		pair := envs[0]
		if pair[0] != key {
			t.Error("Expected ", key, " got ", pair[0])
		}
		
		if pair[1] != value {
			t.Error("Expected ", value, " got ", pair[1])
		}
	}
	
	os.Unsetenv(key)
}
