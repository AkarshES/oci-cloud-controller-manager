FROM odo-docker-signed-local.artifactory.oci.oraclecorp.com/oke/go-boringcrypto-4493-x86_64:1.21.5-243 as builder

ARG COMPONENT
ARG SRC_DIRS
ENV SRC /gopath/src/github.com/oracle/oci-cloud-controller-manager

RUN yum-config-manager --disable \* && yum-config-manager --add-repo https://artifactory.oci.oraclecorp.com/io-ol8-latest-yum-local && yum repolist enabled \
  && yum install -y make \
  && yum clean all \
  && rm -rf /var/lib/yum/* /var/lib/rpm/* /var/cache/yum/* /var/tmp/* /root/.gem /usr/share/doc/*

RUN mkdir -p /go/bin $SRC
ADD . $SRC
WORKDIR $SRC

RUN SRC_DIRS=${SRC_DIRS} make coverage
RUN COMPONENT=${COMPONENT} make clean build

FROM ocr-docker-remote.artifactory.oci.oraclecorp.com/os/oraclelinux:8-slim-fips
RUN rm -f /etc/yum.repos.d/*
COPY artifactory.repo /etc/yum.repos.d/.
RUN microdnf install -y io-ol8-container-hardening

RUN microdnf -y install util-linux e2fsprogs xfsprogs && \
    microdnf update && \
    microdnf clean all && \
    rm -rf /var/cache/yum

COPY scripts/encrypt-mount /sbin/encrypt-mount
COPY scripts/encrypt-umount /sbin/encrypt-umount
COPY scripts/rpm-host /sbin/rpm-host
COPY scripts/chroot-bash /sbin/chroot-bash

RUN chmod 755 /sbin/encrypt-mount
RUN chmod 755 /sbin/encrypt-umount
RUN chmod 755 /sbin/rpm-host
RUN chmod 755 /sbin/chroot-bash

COPY --from=0 /gopath/src/github.com/oracle/oci-cloud-controller-manager/dist/* /usr/local/bin/
