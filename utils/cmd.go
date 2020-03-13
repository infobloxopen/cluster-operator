package utils

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

func GetEnvs(envs []string) [][]string {
	selEnvs := [][]string{}
	if len(envs) > 0 {
		osEnvs := os.Environ()
		for _, e := range osEnvs {
			pair := strings.SplitN(e, "=", 2)
			for _, s := range envs {
				if s == pair[0] {
					selEnvs = append(selEnvs, pair)
				}
			}
		}
	}
	
	return selEnvs
}

func CopyBufferContentsToFile(srcBuff []byte, destFile string) (err error) {
	out, err := os.Create(destFile)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = out.Write(srcBuff); err != nil {
		return
	}
	err = out.Sync()
	return
}

func RunCmd(cmdString string) (*bytes.Buffer, error) {
	var out bytes.Buffer

	cmd := exec.Command("echo", cmdString)
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	var mode os.FileMode = 509
	err = os.MkdirAll("./tmp", mode)
	if err != nil {
		return nil, err
	}

	err = CopyBufferContentsToFile(out.Bytes(), "./tmp/cmd.sh")
	if err != nil {
		return nil, err
	}

	out.Reset()
	cmd = exec.Command("/bin/bash", "./tmp/cmd.sh")
	cmd.Stdout = &out
	var errout bytes.Buffer
	cmd.Stderr = &errout
	err = cmd.Run()
	if err != nil {
		CopyBufferContentsToFile(errout.Bytes(), "./tmp/error.txt")
		return &errout, err
	}

	return &out, nil
}
