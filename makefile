export AWS_ACCESS_KEY_ID	 ?= $(shell aws configure get aws_access_key_id)
export AWS_SECRET_ACCESS_KEY ?= $(shell aws configure get aws_secret_access_key)
export AWS_REGION		     ?= $(shell aws configure get region)
export CLUSTER_OPERATOR_DEVELOPMENT ?= true
export KOPS_CLUSTER_DNS_ZONE ?= soheil.belamaric.com

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

operator-chart:
	helm upgrade -i `cat .id`-cluster-operator --namespace `cat .id` \
		deploy/cluster-operator \
		--set crds.create=true

deploy: .id deploy/cluster.yaml generate operator-chart operator-chart operator-todo

deploy-local: .id deploy/cluster.yaml operator-crds operator-todo

operator-crds:
	kubectl apply -f deploy/cluster-operator/crds/cluster-operator.infobloxopen.github.com_clusters_crd.yaml

operator-todo: .id operator-sdk
	# TODO: move operator-sdk into chart
	OPERATOR_NAME=clusterop .bin/operator-sdk-$(OPERATOR_SDK_VERSION) run --local --namespace `cat .id` --operator-flags='--zap-devel'

operator-debug: .id operator-sdk
	# TODO: move operator-sdk into chart
	OPERATOR_NAME=clusterop .bin/operator-sdk-$(OPERATOR_SDK_VERSION) run --local --namespace `cat .id` --enable-delve

cluster: deploy/cluster.yaml
	# TODO: make our own namespaces
	kubectl create ns `cat .id` || true
	kubectl apply -f deploy/cluster.yaml

status:
	kubectl -n `cat .id` describe cluster example-cluster

delete:
	kubectl -n `cat .id` delete cluster example-cluster

generate:
	operator-sdk generate k8s # codegen

test:
	go build ./...
	git diff --exit-code
