export CLUSTER_OPERATOR_AWS_ACCESS_KEY_ID	 ?= $(shell aws configure get aws_access_key_id)
export CLUSTER_OPERATOR_AWS_SECRET_ACCESS_KEY ?= $(shell aws configure get aws_secret_access_key)
export CLUSTER_OPERATOR_KOPS_STATE_STORE = s3://kops.state.seizadi.infoblox.com
export CLUSTER_OPERATOR_DEVELOPMENT ?= true
export CLUSTER_OPERATOR_REAPER ?= false
export CLUSTER_OPERATOR_KOPS_CLUSTER_DNS_ZONE ?= soheil.belamaric.com

REGISTRY      := infoblox
IMAGE_REPO    := cluster-operator
GIT_COMMIT 	  := $(shell git describe --tags --always || echo pre-commit)
NAMESPACE	  ?= operator
IMAGE         ?= $(GIT_COMMIT)

OPERATOR_SDK_VERSION := v0.15.2
KOPS_VERSION := v1.16.0
export KOPS_PATH ?=".bin/kops"

.id:
	git config user.email | awk -F@ '{print $$1}' > .id

.bin/operator-sdk-$(OPERATOR_SDK_VERSION):
	if [[ ! -f .bin/operator-sdk-$(OPERATOR_SDK_VERSION) ]]; then \
		mkdir -p .bin; \
		curl --fail -Lo $@ https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk-${OPERATOR_SDK_VERSION}-x86_64-$(shell uname -s | tr '[:upper:]' '[:lower:]' | sed 's/darwin/apple-darwin/' | sed 's/linux/linux-gnu/'); \
		chmod +x $@; \
	fi

operator-sdk: .bin/operator-sdk-$(OPERATOR_SDK_VERSION)

$(KOPS_PATH):
	if [[ ! -f $(KOPS_PATH) ]]; then \
		mkdir -p .bin; \
		curl --fail -Lo $@ https://github.com/kubernetes/kops/releases/download/${KOPS_VERSION}/kops-$(shell uname -s | tr '[:upper:]' '[:lower:]')-amd64; \
		chmod +x $@; \
	fi

kops: $(KOPS_PATH)

deploy/cluster.yaml: .id deploy/cluster.yaml.in
	sed "s/{{ .Name }}/`cat .id`/g; s#{{ .sshKey }}#`cat ./ssh/kops.pub`#g" deploy/cluster.yaml.in > $@

operator-chart:
	helm upgrade -i `cat .id`-cluster-operator --namespace `cat .id` \
		deploy/cluster-operator \
		--set crds.create=true

helm-deploy: 
	sed "s/latest/$(IMAGE)/g" deploy/cluster-operator/values.yaml > tmp/values.yaml
	helm template deploy/cluster-operator/. --name phase-1 --namespace $(NAMESPACE) operator -f tmp/values.yaml | kubectl apply -f -

deploy-local: .id deploy/cluster.yaml kops generate operator-crds operator-todo

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

.image-$(IMAGE):
	docker build -t="$(REGISTRY)/$(IMAGE_REPO):$(IMAGE)" .

image: .image-$(IMAGE)

push: image
	docker push $(REGISTRY)/$(IMAGE_REPO):$(IMAGE)


status:
	kubectl -n `cat .id` describe cluster example-cluster

delete:
	kubectl -n `cat .id` delete cluster example-cluster

generate:
	operator-sdk generate k8s # codegen

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

test-vendor: vendor
	[ -z "`git status --porcelain`" ] || { echo "file changes after updating vendoring, check that vendored packages were committed"; exit 1; }

test:
	go build ./...
	git diff --exit-code
