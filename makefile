export AWS_ACCESS_KEY_ID	 ?= $(shell aws configure get aws_access_key_id)
export AWS_SECRET_ACCESS_KEY ?= $(shell aws configure get aws_secret_access_key)
export AWS_REGION		     ?= $(shell aws configure get region)

OPERATOR_SDK_VERSION := v0.15.2

.id:
	git config user.email | awk -F@ '{print $$1}' > .id

.bin/operator-sdk-$(OPERATOR_SDK_VERSION):
	mkdir -p .bin
	curl --fail -Lo $@ https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk-${OPERATOR_SDK_VERSION}-x86_64-$(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/apple-darwin/' | sed 's/linux/linux-gnu/')
	chmod +x $@

operator-sdk: .bin/operator-sdk-$(OPERATOR_SDK_VERSION)

deploy/cluster.yaml: .id deploy/cluster.yaml.in
	sed "s/{{ .Name }}/`cat .id`/g" deploy/cluster.yaml.in > $@

deploy: .id deploy/cluster.yaml generate operator-sdk
	kubectl apply -f deploy/crds/cluster-operator.seizadi.github.com_clusters_crd.yaml
	helm upgrade -i `cat .id` deploy/cluster-operator
	OPERATOR_NAME=clusterop .bin/operator-sdk-$(OPERATOR_SDK_VERSION) run --local --namespace "kops"

cluster: deploy/cluster.yaml
	# TODO: make our own namespaces
	kubectl create ns kops
	kubectl apply -f deploy/cluster.yaml

generate:
	operator-sdk generate k8s # codegen
