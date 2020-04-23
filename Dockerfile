FROM golang:1.14-alpine as builder

ENV OPERATOR_SDK_VERSION=v0.15.2
# RUN mkdir -p .bin

ENV SRC=/go/src/cluster-operator

COPY . ${SRC}
ADD https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk-${OPERATOR_SDK_VERSION}-x86_64-linux-gnu /usr/local/bin/operator-sdk

RUN chmod +x /usr/local/bin/operator-sdk
WORKDIR ${SRC}

RUN operator-sdk generate k8s
RUN go build -o build ./...
RUN ls build

FROM alpine:3.9

ENV SRC=/go/src/cluster-operator
RUN mkdir build
COPY --from=builder ${SRC}/build /

ARG AWSCLI_VERSION=1.16.312
ARG KOPS_VERSION=v1.16.0
ARG KUBECTL_VERSION=v1.16.2
ENV KOPS_PATH=.bin/kops

RUN  apk add --update --no-cache bash python jq ca-certificates groff less \
  && apk add --update --no-cache --virtual build-deps py-pip curl \
  && pip install --upgrade --no-cache-dir awscli==$AWSCLI_VERSION

ADD https://github.com/kubernetes/kops/releases/download/${KOPS_VERSION}/kops-linux-amd64 ${KOPS_PATH}
ADD https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN chmod +x .bin/kops /usr/local/bin/kubectl

RUN chmod +x /manager

ENTRYPOINT ["/manager"]