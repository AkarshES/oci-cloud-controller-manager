#!/bin/bash

set -x -v

#env_file=$(mktemp)
#
#env | grep -v XDG_SESSION_ID | grep -v TEAMCITY_BUILD_PROPERTIES_FILE | grep -v TMPDIR | grep -v JDK_18_x64 | grep -v TEAMCITY_CAPTURE_ENV | grep -v "^USER" | grep -v TEMP | grep -v JDK_18 | grep -v JRE_HOME | grep -v PATH | grep -v PWD | grep -v JAVA_HOME | grep -v LANG | grep -v SHLVL | grep -v HOME | grep -v JDK_HOME | grep -v XDG_RUNTIME_DIR > "$env_file"

>.env_file
for var in $(compgen -v | grep -Ev '^(BASH)'); do
    var_fixed=$(printf "%s" "${!var}" | tr -d '\n' )
    echo "$var=${var_fixed}" >>.env_file
done

echo -e "Env:\n" ${env_file}

#docker run \
#	--volumes-from "${DIND_NAME}" \
#	--env-file "${env_file}" \
#	-e LOCAL_RUN=${LOCAL_RUN:-0} \
#	-e PATH="/usr/local/go/bin:/gopath/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/usr/local/oci/bin" \
#	-e GOPATH="/gopath" \
#	-e http_proxy=http://10.68.69.53:80 \
#	-e https_proxy=http://10.68.69.53:80 \
#	--net=host \
#	-v "$(pwd)/e2e-tests":/gopath/src/bitbucket.oci.oraclecorp.com/oke/e2e-tests \
#	-w /gopath/src/bitbucket.oci.oraclecorp.com/oke/e2e-tests \
#	-v "$(pwd)/config":/config \
#	-v "$(pwd)/secrets":/secrets \
#	${E2E_TEST_BASE_IMAGE:-odo-docker-signed-local.artifactory.oci.oraclecorp.com/odx-oke/oke/k8-manager-base:go1.20.4-1.0.19} \
#	/bin/bash -c "yum install -yy openssl && make \"$@\""

docker --config "$DOCKER_CONFIG_DIR" run \
	--volumes-from "${DIND_NAME}" \
	--env-file "${env_file}" \
	-e LOCAL_RUN="${LOCAL_RUN:-0}" \
	-e PATH="/usr/local/go/bin:/gopath/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/usr/local/oci/bin" \
	-e GOPATH="/gopath" \
	-e http_proxy=http://10.68.69.53:80 \
	-e https_proxy=http://10.68.69.53:80 \
	--net=host \
	-v "$(pwd)/..":/gopath/src/github.com/oracle/oci-cloud-controller-manager  \
	-w /gopath/src/github.com/oracle/oci-cloud-controller-manager  \
	-v "$(pwd)/../config":/config \
	-v "$(pwd)/../secrets":/secrets \
	"${E2E_TEST_BASE_IMAGE:-odo-docker-signed-local.artifactory.oci.oraclecorp.com/odx-oke/oke/k8-manager-base:ginkgo-1.0.9}" \
	/bin/bash -c "ls -ltr && make \"$1\""
