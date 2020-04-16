FROM golang:1.14-alpine as builder

# RUN apk --update add --no-cache --virtual .build-deps \
#     gcc libc-dev linux-headers git

ENV SRC=/go/src/cluster-operator
COPY . ${SRC}
WORKDIR ${SRC}
RUN go build -o deploy ./...

# FROM alpine:3.5

# ENV AWS_CLI_VERSION 1.11.131

# RUN apk --no-cache update && \
#    apk --no-cache add python openldap-clients py-pip py-setuptools ca-certificates groff less && \
#    pip --no-cache-dir install awscli==${AWS_CLI_VERSION} && \
#    rm -rf /var/cache/apk/*

# RUN pip install --no-cache-dir awscli --upgrade --user
# RUN /root/.local/bin/aws --version
# ENV SRC=/go/src/github.com/infoblox-cto/space-controller
# COPY --from=builder ${SRC}/bin/ /

# Allows to verify certificates
# RUN apk update --no-cache && apk add --no-cache ca-certificates git curl
ENTRYPOINT ["/manager"]
# CMD ["sleep 10000"]

# sudo docker run -d --name space -v "/etc/certs/:/etc/certs/" infobloxcto/space-controller:latest -kubeconfig=/etc/certs/kubeconfig