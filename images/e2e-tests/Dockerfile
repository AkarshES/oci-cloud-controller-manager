FROM ocr-docker-remote.artifactory.oci.oraclecorp.com/os/oraclelinux:7-slim

#install latest oci cli from yum repo
#ARG OCI_CLI_VERSION
ARG KUBECTL_VERSION

#enable devo repo for oci-cli installation. oci-cli installation will have python3 installed.
#reinstall glibc-common to populate /usr/share/i18n/charmap so lcoaledef can work
#TODO confirm whether python3-distutils is needed.
#TODO I don't think OCL_CLI_VERSION matters so python36-oci-cli should be sufficient.
RUN yum install -y oraclelinux-developer-release-el7 oracle-epel-release-el7 && \
    yum install -y python36-oci-cli openssl go && \
    yum reinstall -y glibc-common && \
    yum clean all

#set locale
RUN localedef -i C -f UTF-8 C.UTF-8

ENV LC_ALL="C.UTF-8"
ENV LANG="C.UTF-8"

# Install test runner
RUN go get github.com/onsi/ginkgo/ginkgo

# Install kubectl
RUN curl -LO https://artifactory.oci.oraclecorp.com:443/build-service-test-generic-local/oci-cloud-controller-manager/kubectl/kubectl-${KUBECTL_VERSION}/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl

# Install sops to decrypt secrets
RUN curl -L https://artifactory.oci.oraclecorp.com/atd-release-generic-local/sops/refs/heads/master/7/sops-linux-amd64 > /usr/local/bin/sops && \
    chmod +x /usr/local/bin/sops


RUN mkdir -p $GOPATH/src/github.com/oracle/oci-cloud-controller-manager
COPY . $GOPATH/src/github.com/oracle/oci-cloud-controller-manager/
RUN chmod +x $GOPATH/src/github.com/oracle/oci-cloud-controller-manager/images/e2e-tests/e2e_pop_run.sh

RUN mv $GOPATH/src/github.com/oracle/oci-cloud-controller-manager/images/e2e-tests/create_oci_config.sh /usr/local/bin/create_oci_config.sh && chmod +x /usr/local/bin/create_oci_config.sh
