export CLUSTER_OPERATOR_AWS_ACCESS_KEY_ID	 ?= $(shell aws configure get aws_access_key_id)
export CLUSTER_OPERATOR_AWS_SECRET_ACCESS_KEY ?= $(shell aws configure get aws_secret_access_key)
export CLUSTER_OPERATOR_AWS_REGION ?= us-east-2

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


operator-chart:
	helm upgrade -i `cat .id`-cluster-operator --namespace `cat .id` \
		deploy/cluster-operator \
		--set crds.create=true

helm-deploy: 
	sed "s/latest/$(IMAGE)/g" deploy/cluster-operator/values.yaml > tmp/values.yaml
	helm template deploy/cluster-operator/. --name phase-1 --namespace $(NAMESPACE) operator -f tmp/values.yaml | kubectl apply -f -

namespace:
	kubectl create ns $(NAMESPACE) || true

cert-manager: 
	kubectl apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.14.3/cert-manager.yaml

docker-local:
	docker build -t="$(REGISTRY)/$(IMAGE_REPO):$(IMAGE)" .
	kind load docker-image $(REGISTRY)/$(IMAGE_REPO):$(IMAGE)

deploy-local: namespace
	helm template deploy/cluster-operator/. --name phase-1 --namespace $(NAMESPACE) operator -f tmp/values.yaml \
	--set image.tag=$(IMAGE) --set image.pullPolicy=Never --set local=true --set aws.secretKey=$(CLUSTER_OPERATOR_AWS_SECRET_ACCESS_KEY) \
	--set aws.keyID=$(CLUSTER_OPERATOR_AWS_ACCESS_KEY_ID) --set aws.region=$(CLUSTER_OPERATOR_AWS_REGION)| kubectl apply -f -

operator-crds:
	kubectl apply -f deploy/cluster-operator/crds/cluster-operator.infobloxopen.github.com_clusters_crd.yaml

operator-todo: .id operator-sdk
	# TODO: move operator-sdk into chart
	OPERATOR_NAME=clusterop .bin/operator-sdk-$(OPERATOR_SDK_VERSION) run --local --namespace `cat .id` --operator-flags='--zap-devel'

operator-debug: .id operator-sdk
	# TODO: move operator-sdk into chart
	OPERATOR_NAME=clusterop .bin/operator-sdk-$(OPERATOR_SDK_VERSION) run --local --namespace `cat .id` --enable-delve


deploy/cluster.yaml: .id deploy/cluster.yaml.in
	sed "s/{{ .Name }}/$(NAMESPACE)/g; s#{{ .sshKey }}#`cat ./ssh/kops.pub`#g" deploy/cluster.yaml.in > $@

cluster: deploy/cluster.yaml namespace
	kubectl apply -f deploy/cluster.yaml

.image-$(IMAGE):
	docker build -t="$(REGISTRY)/$(IMAGE_REPO):$(IMAGE)" .

image: .image-$(IMAGE)

push: image
	docker push $(REGISTRY)/$(IMAGE_REPO):$(IMAGE)

status:
	kubectl -n $(NAMESPACE) describe cluster example-cluster

delete:
	kubectl -n $(NAMESPACE) delete cluster example-cluster

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
