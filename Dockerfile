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
ENV WATCH_NAMESPACE=hryan
ENV KOPS_STATE_STORE=s3://kops.state.seizadi.infoblox.com
ENV AWS_ACCESS_KEY_ID=AEJ9IveryfakehF728
ENV AWS_SECRET_ACCESS_KEY=0vcAYtBJEhisisfakeMDk4V5MqfUtUnH
RUN mkdir build
COPY --from=builder ${SRC}/build /

COPY --from=builder ${SRC}/config config
RUN ls
RUN chmod +x /manager

ENTRYPOINT ["/manager"]