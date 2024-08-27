#!/bin/bash
# Note: Not using Docker in Docker anymore

set -x -v

# Remove usage of openssl since its not necessary.
# TODO: make appropriate changes to "hack/run_e2e_test.sh" in the active release branches
sed -i'.bak' -e "s/sed 's\/ \/\/g' | openssl enc -base64 -d -A/base64 -d/" hack/run_e2e_test.sh

# Convert environment variables with "nil" as value to empty values
sed -i'.bak' -e "s/=nil/=/" "${env_file}"

# Pipelines does not allow parameters starting with "OCI_*"
# Convert our non credential E2E parameters which coincidentally start with "OCI_"
# from "OOCI_" (hack to use Pipeline parameters) to "OCI_" so that E2E shell script does not need to be updated.
sed -i'.bak' -e "s/^OOCI_/OCI_/" "${env_file}"

# Check disk usage
sudo df -h

# Run make command within docker container to achieve independence from Runner Instance architecture
docker --config="$DOCKER_CONFIG_DIR" run \
	--env-file="${env_file}" \
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
	"${E2E_TEST_BASE_IMAGE:-odo-docker-signed-local.artifactory.oci.oraclecorp.com/odx-oke/oke/k8-manager-base:ginkgo-1.0.9}" \
	/bin/bash -c "ls -ltr && env && ./create_oci_config.sh && make \"$1\""
