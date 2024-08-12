#!/bin/bash

set -x -v

env_file=$(mktemp)
unset OCI_RESOURCE_PRINCIPAL_PRIVATE_PEM
# env | grep -v XDG_SESSION_ID | grep -v TEAMCITY_BUILD_PROPERTIES_FILE | grep -v TMPDIR | grep -v JDK_18_x64 | grep -v TEAMCITY_CAPTURE_ENV | grep -v "^USER" | grep -v TEMP | grep -v JDK_18 | grep -v JRE_HOME | grep -v PATH | grep -v PWD | grep -v JAVA_HOME | grep -v LANG | grep -v SHLVL | grep -v HOME | grep -v JDK_HOME | grep -v XDG_RUNTIME_DIR > "$env_file"

env > "${env_file}"

>.env_file
for var in $(compgen -v | grep -Ev '^(BASH)'); do
    var_fixed=$(printf "%s" "${!var}" | tr -d '\n' )
    echo "$var=${var_fixed}" >>.env_file
done

export E2E_TEST_BASE_IMAGE=${E2E_TEST_BASE_IMAGE//artifactory.oci.oraclecorp.com/pipelines.artifactory.us-phoenix-1.oci.oracleiaas.com}

#docker --config "$DOCKER_CONFIG_DIR" run \
#	--volumes-from "${DIND_NAME}" \
#	--env-file "${env_file}" \
#	-e LOCAL_RUN="${LOCAL_RUN:-0}" \
#	-e PATH="/usr/local/go/bin:/gopath/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/usr/local/oci/bin" \
#	-e GOPATH="/gopath" \
#	-e http_proxy="${HTTP_PROXY}" \
#	-e https_proxy="${HTTPS_PROXY}" \
#	-e no_proxy="${NO_PROXY}" \
#	--net=host \
#	-v "$(pwd)":/gopath/src/github.com/oracle/oci-cloud-controller-manager  \
#	-w /gopath/src/github.com/oracle/oci-cloud-controller-manager  \
#	-v "$(pwd)/config":/config \
#	-v "$(pwd)/secrets":/secrets \
#	"${E2E_TEST_BASE_IMAGE}" \
#	/bin/bash -c "yum install -yy openssl && go version && make \"$1\""

docker --config "$DOCKER_CONFIG_DIR" run \
	--volumes-from "${DIND_NAME}" \
	--env-file "${env_file}" \
	-e LOCAL_RUN="${LOCAL_RUN:-0}" \
	-e PATH="/usr/local/go/bin:/gopath/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/usr/local/oci/bin" \
	-e GOPATH="/gopath" \
	-e http_proxy="${HTTP_PROXY}" \
	-e https_proxy="${HTTPS_PROXY}" \
	-e no_proxy="${NO_PROXY}" \
	--net=host \
	-v "$(pwd)":/gopath/src/github.com/oracle/oci-cloud-controller-manager  \
	-w /gopath/src/github.com/oracle/oci-cloud-controller-manager  \
	-v "$(pwd)/config":/config \
	-v "$(pwd)/secrets":/secrets \
	"${E2E_TEST_BASE_IMAGE:-odo-docker-signed-local.pipelines.artifactory.us-phoenix-1.oci.oracleiaas.com/odx-oke/oke/k8-manager-base:ginkgo-1.0.9}" \
	/bin/bash -c "yum install -yy openssl && go version && make \"$1\""
