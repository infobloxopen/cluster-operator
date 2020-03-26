/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package coredns is the implementation of pkg/dnsprovider interface for CoreDNS
package coredns

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	etcdc "github.com/coreos/etcd/client"
	gcfg "gopkg.in/gcfg.v1"
	"k8s.io/klog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

// "coredns" should be used to use this DNS provider
const (
	ProviderName = "coredns"
)

// Config to override defaults
type Config struct {
	Global struct {
		EtcdEndpoints    string `gcfg:"etcd-endpoints"`
		DNSZones         string `gcfg:"zones"`
		CoreDNSEndpoints string `gcfg:"coredns-endpoints"`
	}
}

func init() {
	dnsprovider.RegisterDnsProvider(ProviderName, func(config io.Reader) (dnsprovider.Interface, error) {
		return newCoreDNSProviderInterface(config)
	})
}

// newCoreDnsProviderInterface creates a new instance of an CoreDNS DNS Interface.
func newCoreDNSProviderInterface(config io.Reader) (*Interface, error) {
	etcdEndpoints := "http://federation-dns-server-etcd:2379"
	etcdPathPrefix := "skydns"
	dnsZones := ""

	// Possibly override defaults with config below
	if config != nil {
		var cfg Config
		if err := gcfg.ReadInto(&cfg, config); err != nil {
			klog.Errorf("Couldn't read config: %v", err)
			return nil, err
		}
		etcdEndpoints = cfg.Global.EtcdEndpoints
		dnsZones = cfg.Global.DNSZones
	}
	klog.Infof("Using CoreDNS DNS provider")

	if dnsZones == "" {
		return nil, fmt.Errorf("need to provide at least one DNS Zone")
	}

	etcdCfg := etcdc.Config{
		Endpoints: strings.Split(etcdEndpoints, ","),
		Transport: etcdc.DefaultTransport,
	}

	c, err := etcdc.New(etcdCfg)
	if err != nil {
		return nil, fmt.Errorf("create etcd client from the config failed")
	}
	etcdKeysAPI := etcdc.NewKeysAPI(c)

	intf := newInterfaceWithStub(etcdKeysAPI)
	intf.etcdPathPrefix = etcdPathPrefix
	zoneList := strings.Split(dnsZones, ",")

	intf.zones = Zones{intf: intf}
	for index, zoneName := range zoneList {
		zone := Zone{domain: zoneName, id: strconv.Itoa(index), zones: &intf.zones}
		intf.zones.zoneList = append(intf.zones.zoneList, zone)
	}

	return intf, nil
}
