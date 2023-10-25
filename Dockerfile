FROM odo-docker-signed-local.artifactory.oci.oraclecorp.com/oke/go-boringcrypto-4493-x86_64:1.19.13-188 as builder

ARG COMPONENT
ARG SRC_DIRS

ENV SRC /gopath/src/github.com/oracle/oci-cloud-controller-manager

RUN mkdir -p /go/bin $SRC
ADD . $SRC
WORKDIR $SRC

RUN SRC_DIRS=${SRC_DIRS} make coverage
RUN COMPONENT=${COMPONENT} make clean build

FROM ocr-docker-remote.artifactory.oci.oraclecorp.com/os/oraclelinux:7-slim

RUN yum-config-manager --disable \* && yum-config-manager --add-repo https://artifactory.oci.oraclecorp.com/io-ol7-latest-yum-local && yum repolist enabled \
  && yum install -y util-linux e2fsprogs \
  && yum install -y xfsprogs \
  && yum update -y \
  && rm -rf /var/cache/yum

COPY scripts/encrypt-mount /sbin/encrypt-mount
COPY scripts/encrypt-umount /sbin/encrypt-umount
COPY scripts/rpm-host /sbin/rpm-host
RUN chmod 755 /sbin/encrypt-mount
RUN chmod 755 /sbin/encrypt-umount
RUN chmod 755 /sbin/rpm-host

COPY --from=0 /gopath/src/github.com/oracle/oci-cloud-controller-manager/dist/* /usr/local/bin/
