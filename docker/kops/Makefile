IMAGE_VERSION  ?= v1.16.0-1.0
IMAGE_REGISTRY ?= infoblox
SERVER_IMAGE   := $(IMAGE_REGISTRY)/kops
DOCKERFILE_PATH := $(CURDIR)
SERVER_DOCKERFILE := $(DOCKERFILE_PATH)/Dockerfile

.PHONY: docker
docker:
	@docker build -f $(SERVER_DOCKERFILE) -t $(SERVER_IMAGE):$(IMAGE_VERSION) .
	@docker tag $(SERVER_IMAGE):$(IMAGE_VERSION) $(SERVER_IMAGE):latest
	@docker image prune -f --filter label=stage=server-intermediate
	@docker push $(SERVER_IMAGE)
.PHONY: push
push:
	@docker push $(SERVER_IMAGE)
