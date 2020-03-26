module github.com/infobloxopen/cluster-operator

go 1.13

require (
	github.com/aws/aws-sdk-go v1.29.21
	github.com/operator-framework/operator-sdk v0.15.2
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	golang.org/x/tools v0.0.0-20191203134012-c197fd4bf371 // indirect
	gopkg.in/yaml.v2 v2.2.7
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kops v1.18.0-alpha.2
	k8s.io/kubectl v0.0.0
	k8s.io/utils v0.0.0-20200229041039-0a110f9eb7ab // indirect
	sigs.k8s.io/controller-runtime v0.4.0
)

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309 // Required by Helm

replace github.com/openshift/api => github.com/openshift/api v0.0.0-20190924102528-32369d4db2ad // Required until https://github.com/operator-framework/operator-lifecycle-manager/pull/1241 is resolved

replace k8s.io/api => k8s.io/api v0.0.0-20191114100352-16d7abae0d2a

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20191114101535-6c5935290e33

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20191114112024-4bbba8331835

replace k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191114113550-6123e1c827f7

// We need a newer component-base
//  replace k8s.io/component-base => k8s.io/component-base kubernetes-1.17.0-rc.2
replace k8s.io/component-base => k8s.io/component-base v0.0.0-20191204084121-18d14e17701e

// Dependencies we don't really need, except that kubernetes specifies them as v0.0.0 which confuses go.mod
//replace k8s.io/apiserver => k8s.io/apiserver kubernetes-1.16.3
//replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver kubernetes-1.16.3
//replace k8s.io/kube-scheduler => k8s.io/kube-scheduler kubernetes-1.16.3
//replace k8s.io/kube-proxy => k8s.io/kube-proxy kubernetes-1.16.3
//replace k8s.io/cri-api => k8s.io/cri-api kubernetes-1.16.3
//replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib kubernetes-1.16.3
//replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers kubernetes-1.16.3
//replace k8s.io/component-base => k8s.io/component-base kubernetes-1.16.3
//replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap kubernetes-1.16.3
//replace k8s.io/metrics => k8s.io/metrics kubernetes-1.16.3
//replace k8s.io/sample-apiserver => k8s.io/sample-apiserver kubernetes-1.16.3
//replace k8s.io/kube-aggregator => k8s.io/kube-aggregator kubernetes-1.16.3
//replace k8s.io/kubelet => k8s.io/kubelet kubernetes-1.16.3
//replace k8s.io/cli-runtime => k8s.io/cli-runtime kubernetes-1.16.3
//replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager kubernetes-1.16.3
//replace k8s.io/code-generator => k8s.io/code-generator kubernetes-1.16.3

replace k8s.io/apiserver => k8s.io/apiserver v0.0.0-20191114103151-9ca1dc586682

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191114105449-027877536833

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20191114111229-2e90afcb56c7

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20191114110717-50a77e50d7d9

replace k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190828162817-608eb1dad4ac

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20191114112310-0da609c4ca2d

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20191114112655-db9be3e678bb

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20191114111741-81bb9acf592d

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20191114105837-a4a2842dc51b

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20191114104439-68caf20693ac

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20191114103820-f023614fb9ea

replace k8s.io/kubelet => k8s.io/kubelet v0.0.0-20191114110954-d67a8e7e2200

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191114110141-0a35778df828

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20191114111510-6d1ed697a64b

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20191004115455-8e001e5d1894

require (
	cloud.google.com/go v0.38.0 // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/Masterminds/sprig v2.17.1+incompatible // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/bazelbuild/bazel-gazelle v0.19.1 // indirect
	github.com/blang/semver v3.5.1+incompatible // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/client9/misspell v0.3.4 // indirect
	github.com/coreos/etcd v3.3.17+incompatible // indirect
	github.com/denverdino/aliyungo v0.0.0-20191128015008-acd8035bbb1d // indirect
	github.com/digitalocean/godo v1.19.0 // indirect
	github.com/docker/engine-api v0.0.0-20160509170047-dea108d3aa0c // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20180422025557-ae226422660e // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/go-bindata/go-bindata v3.1.2+incompatible // indirect
	github.com/go-ini/ini v1.51.0 // indirect
	github.com/go-logr/logr v0.1.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/gophercloud/gophercloud v0.7.1-0.20200116011225-46fdd1830e9a // indirect
	github.com/gorilla/mux v1.7.1 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/jacksontj/memberlistmesh v0.0.0-20190905163944-93462b9d2bb7 // indirect
	github.com/jpillora/backoff v0.0.0-20170918002102-8eab2debe79d // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/miekg/coredns v0.0.0-20161111164017-20e25559d5ea // indirect
	github.com/miekg/dns v1.1.4 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/sftp v0.0.0-20160930220758-4d0e916071f6 // indirect
	github.com/prometheus/client_golang v1.2.1 // indirect
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/spotinst/spotinst-sdk-go v1.43.0 // indirect
	github.com/stretchr/testify v1.4.0 // indirect
	github.com/urfave/cli v1.20.0 // indirect
	github.com/vmware/govmomi v0.20.1 // indirect
	github.com/weaveworks/mesh v0.0.0-20170419100114-1f158d31de55 // indirect
	go.uber.org/zap v1.10.0 // indirect
	golang.org/x/crypto v0.0.0-20191202143827-86a70503ff7e // indirect
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20191128015809-6d18c012aee9 // indirect
	google.golang.org/api v0.17.0 // indirect
	gopkg.in/gcfg.v1 v1.2.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	honnef.co/go/tools v0.0.1-2019.2.3 // indirect
	k8s.io/cli-runtime v0.0.0
	k8s.io/component-base v0.0.0 // indirect
	k8s.io/helm v2.16.1+incompatible // indirect
	k8s.io/legacy-cloud-providers v0.0.0 // indirect
	sigs.k8s.io/controller-tools v0.2.4 // indirect
	sigs.k8s.io/yaml v1.1.0 // indirect
)

require github.com/Azure/go-autorest v12.2.0+incompatible // indirect
