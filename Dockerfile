
FROM odo-docker-signed-local.artifactory.oci.oraclecorp.com/odx-oke/oke/golang-buildbox:1.12.7-fips-471967d48d010fd56ace221381853f4709484717-40 as builder

ARG COMPONENT

ENV SRC /go/src/github.com/oracle/oci-cloud-controller-manager

ENV GOPATH /go/
RUN mkdir -p /go/bin $SRC
ADD . $SRC
WORKDIR $SRC

RUN COMPONENT=${COMPONENT} make clean build

FROM ocr-docker-remote.artifactory.oci.oraclecorp.com/os/oraclelinux:7-slim

RUN yum-config-manager --disable \* && yum-config-manager --add-repo https://artifactory.oci.oraclecorp.com/io-ol7-latest-yum-local && yum repolist enabled \
  && yum install -y util-linux e2fsprogs \
  && rm -rf /var/cache/yum

COPY --from=0 /go/src/github.com/oracle/oci-cloud-controller-manager/dist/* /usr/local/bin/
