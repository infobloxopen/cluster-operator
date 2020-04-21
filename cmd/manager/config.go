package main

import (
	"github.com/spf13/pflag"
)

const (
	// Change below variables to serve metrics on different host or port.
	defaultTmpDir = "/tmp"

	//Kops
	defaultKopsStateStore     = "s3://kops.state.seizadi.infoblox.com"
	defaultKopsClusterDnsZone = "soheil.belamaric.com"
	defaultSSHKey             = "kops.pub"
	defaultKopsContainer = "soheileizadi/kops:v1.0"
	defaultKopsKubeDir = "/kube"

	//Docker
	defaultDockerBinPath = "/usr/local/bin/docker"

	//Metrics
	defaultMetricsHost               = "0.0.0.0"
	defaultMetricsPort         int32 = 8383
	defaultOperatorMetricsPort int32 = 8686

	// Configuration
	defaultConfigDirectory = "deploy/"
	defaultConfigFile      = ""
	appEnvPrefix           = "CLUSTER_OPERATOR"

	// Critical values that are imported from eventual configMap using environment variables.
	// The current list that must be supported currently managed from Makefile for development.
	// AWS Cloud Access
	defaultAwsAccessKeyID     = ""
	defaultAwsSecretAccessKey = ""
	defaultAwsRegion          = ""

	// Development Flag
	defaultClusterOperatorDevelopment bool = false
)

var (
	// define flag overrides
	flagTmpDir = pflag.String("tmp.dir", defaultTmpDir, "temp directory")

	// Kops
	flagKopsStateStore     = pflag.String("kops.state.store", defaultKopsStateStore, "kops state store")
	flagKopsClusterDnsZone = pflag.String("kops.cluster.dns.zone", defaultKopsClusterDnsZone, "kops cluster DNS zone")
	flagSSHKey             = pflag.String("kops.ssh.key", defaultSSHKey, "kops ssh key")
	flagKopsContainer = pflag.String("kops.container", defaultKopsContainer, "kops container")
	flagKopsKubeDir  = pflag.String("kops.kube.dir", defaultKopsKubeDir, "kops kube directory")

	//Docker
	flagDockerBinPath = pflag.String("docker.bin.path", defaultDockerBinPath, "docker bin path")

	// Metrics
	flagMetricsHost         = pflag.String("metrics.host", defaultMetricsHost, "host of metrics")
	flagMetricsPort         = pflag.Int32("metrics.port", defaultMetricsPort, "port of metrics")
	flagOperatorMetricsPort = pflag.Int32("operator.metrics.port", defaultOperatorMetricsPort, "port of metrics operator")

	// Configuration
	flagConfigDirectory = pflag.String("config.source", defaultConfigDirectory, "directory of the configuration file")
	flagConfigFile      = pflag.String("config.file", defaultConfigFile, "directory of the configuration file")

	// AWS Cloud Access
	flagAwsAccessKeyID     = pflag.String("aws.access.key.id", defaultAwsAccessKeyID, "AWS access key ID")
	flagAwsSecretAccessKey = pflag.String("aws.secret.access.key", defaultAwsSecretAccessKey, "AWS secret access key")
	flagAwsRegion          = pflag.String("aws.region", defaultAwsRegion, "AWS region")

	// Developement Flag so eliminate Cloud interaction and speedup local developement
	flagClusterOperatorDevelopment = pflag.Bool("development", defaultClusterOperatorDevelopment, "cluster operator development")
)
