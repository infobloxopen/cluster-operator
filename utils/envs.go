package utils

import (
	"github.com/spf13/viper"
	"os"
	"strings"
)

func GetEnvs(filterEnvs []string) [][]string {
	selEnvs := [][]string{}
	if len(filterEnvs) > 0 {
		osEnvs := os.Environ()
		for _, e := range osEnvs {
			pair := strings.SplitN(e, "=", 2)
			for _, s := range filterEnvs {
				if s == pair[0] {
					selEnvs = append(selEnvs, pair)
					break
				}
			}
		}
	}

	return selEnvs
}

func CheckEnvs(envs [][]string, reqEnvs []string) []string {
	missingEnvs := []string{}
	for _, e := range reqEnvs {
		found := false
		for _, pair := range envs {
			if e == pair[0] {
				found = true
				break
			}
		}
		if found == false {
			missingEnvs = append(missingEnvs, e)
		}
	}
	return missingEnvs
}

func GetDockerEnvFlags() string {
	envFlags := " -e AWS_ACCESS_KEY_ID=" + viper.GetString("aws.access.key.id") +
		"-e AWS_SECRET_ACCESS_KEY=" + viper.GetString("aws.secret.access.key") +
		"-e KOPS_STATE_STORE=" + viper.GetString("kops.state.store")
	return envFlags
}
