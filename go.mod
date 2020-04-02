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
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32
	github.com/spotinst/spotinst-sdk-go v1.43.0 // indirect
	google.golang.org/api v0.17.0 // indirect
	k8s.io/cli-runtime v0.0.0
)
