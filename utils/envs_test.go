package utils

import (
	"os"
	"testing"
)

func TestGetEnvs(t *testing.T) {
	key := "CLUSTER_OPERATOR_TEST_9898"
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

func TestCheckEnvs(t *testing.T) {
	key := "CLUSTER_OPERATOR_TEST_9898"
	value := "2020"
	missingKeys := []string{"CLUSTER_OPERATOR_TEST_NOT_IN_ENV1", "CLUSTER_OPERATOR_TEST_NOT_IN_ENV2"}
	os.Setenv(key, value)
	envs := GetEnvs([]string{key, missingKeys[0]})
	missingEnvs := CheckEnvs(envs, missingKeys)
	
	if len(missingEnvs) != len(missingKeys) {
		t.Error("Expected 2 got ", len(envs))
	} else {
		count :=0
		for _, e := range missingEnvs {
			for _, m := range missingKeys {
				if e == m {
					count += 1
				}
			}
		}
		if count != len(missingKeys) {
			t.Error("Expected ", len(missingKeys), " missingKeys matches, got ", count)
		}
	}
	os.Unsetenv(key)
}
