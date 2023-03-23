FROM odo-docker-signed-local.artifactory.oci.oraclecorp.com/odx-oke/oke/golang-buildbox:1.19.1-fips-9b221ddd244a2b114c9a1053d09993fb4f9ee526-66 as builder

ARG COMPONENT

ENV SRC /gopath/src/github.com/oracle/oci-cloud-controller-manager

RUN mkdir -p /go/bin $SRC
ADD . $SRC
WORKDIR $SRC

RUN COMPONENT=${COMPONENT} make clean build

FROM ocr-docker-remote.artifactory.oci.oraclecorp.com/os/oraclelinux:7-slim

RUN yum-config-manager --disable \* && yum-config-manager --add-repo https://artifactory.oci.oraclecorp.com/io-ol7-latest-yum-local && yum repolist enabled

RUN yum install -y util-linux \
  && yum install -y e2fsprogs \
  && yum install -y xfsprogs \
  && yum clean all

COPY --from=0 /gopath/src/github.com/oracle/oci-cloud-controller-manager/dist/* /usr/local/bin/
