package main

import (
	"github.com/spf13/viper"
	"os"
	"testing"
)

func TestConfig1(t *testing.T) {
	host := "info"
	os.Setenv(appEnvPrefix+"_METRICS_HOST", host)
	if viper.GetString("metrics.host") != host {
		t.Error("Expected", host, "Got", viper.GetString("metrics.host"))
	}
}

func TestConfig2(t *testing.T) {
	var port int32 = 42
	os.Setenv("VIPER_METRICS_PORT", "42")
	if viper.GetInt32("metrics.port") != port {
		t.Error("Expected", port, "Got", viper.GetString("metrics.port"))
	}
}
