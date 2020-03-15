package utils

import (
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
					break;
				}
			}
		}
	}
	
	return selEnvs
}

func CheckEnvs(envs [][]string, reqEnvs[]string) []string {
	missingEnvs := []string{}
	for _, e := range reqEnvs {
		found := false
		for _, pair := range envs {
			if e == pair[0] {
				found = true
				break;
			}
		}
		if found == false {
			missingEnvs = append(missingEnvs, e)
		}
	}
	return missingEnvs
}
