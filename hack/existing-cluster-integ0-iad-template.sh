#!/bin/bash

##################################################################################################
# This template can be used to tweak the environment variables needed to run the E2E tests locally #
# Default behavior:
# Runs test on an existing cluster in dev0-iad

# To run the tests:
# 1. Change the FOCUS variable here to specify the subset of E2E tests to run
# 2. Set CLUSTER_KUBECONFIG and CLOUD_CONFIG if needed
# 3. run 'source existing-cluster-dev0-env-template.sh' to set the variables
# 4. run 'make run-ccm-e2e-tests-local`
##################################################################################################

# The test suites to run (can replace or add tags)
export FOCUS="\[managedNsg\]"

# This variable tells the test not to install oci cli and wipe out your .oci/config
export LOCAL_RUN=1
export TC_BUILD=0

# This allows you to use your existing cluster
export ENABLE_CREATE_CLUSTER=false

# Set path to kubeconfig of existing cluster if it does not exist in default path. Defaults to $HOME/.kube/config_*
export CLUSTER_KUBECONFIG_AMD=$HOME/.kube/config
export CLUSTER_KUBECONFIG_ARM=$HOME/.kube/config

# Set path to cloud_config of existing cluster if it does not exist in default path. Defaults to $HOME/cloudconfig_*
export CLOUD_CONFIG_AMD=$HOME/go-workspace/src/bitbucket.oci.oraclecorp.com/oke/oci-cloud-controller-manager/hack/cloudconfig_integ_iad
export CLOUD_CONFIG_ARM=$HOME/go-workspace/src/bitbucket.oci.oraclecorp.com/oke/oci-cloud-controller-manager/hack/cloudconfig_integ_iad


export IMAGE_PULL_REPO="iad.ocir.io/okedev/e2e-tests/"
export ADLOCATION="zkJl:US-ASHBURN-AD-1"


#KMS key for CMEK testing
export CMEK_KMS_KEY="ocid1.key.oc1.iad.bbpqrkz5aaeuk.abuwcljtw6qdzog4gpw5pohxrdztnwthvbwu7kdt3q52tc7snaxph3byqpya"

#NSG Network security group created in cluster's VCN
export NSG_OCIDS=""

#Reserved IP created in e2e test compartment
export RESERVED_IP="129.159.98.123"

#Architecture to run tests on
export ARCHITECTURE_AMD="AMD"
export ARCHITECTURE_ARM="ARM"

#Focus the tests : ARM, AMD or BOTH
export SCOPE="ARM"

#NSG Network security group created in cluster's VCN for backend management, this NSG will have to be attached to the nodes manually for tests to pass
export BACKEND_NSG_OCIDS="ocid1.networksecuritygroup.oc1.iad.aaaaaaaal6345wfpvueik24f3piwiey5z2g4ifbexpknynxt4abds56qnjsa"
# For debugging the tests in existing cluster, do not turn it off by default.
# export DELETE_NAMESPACE=false


export HTTP_PROXY=http://www-proxy-idc.in.oracle.com:80
export HTTPS_PROXY=http://www-proxy-idc.in.oracle.com:80


# FSS volume handle
# format is FileSystemOCID:serverIP:path

export FSS_VOLUME_HANDLE="ocid1.filesystem.oc1.iad.aaaaaaaaaabwoogpnfqwillqojxwiotjmfsc2ylefuyqaaaa:10.0.0.4:/FileSystem-Fss-Dyn-E2e"
export FSS_VOLUME_HANDLE_ARM="ocid1.filesystem.oc1.iad.aaaaaaaaaabwooosnfqwillqojxwiotjmfsc2ylefuyqaaaa:10.0.0.4:/FileSystem-Fss-Dyn-E2E-Arm"

export MNT_TARGET_ID="ocid1.mounttarget.oc1.iad.aaaaacvippy7r7n4nfqwillqojxwiotjmfsc2ylefuyqaaaa"
export MNT_TARGET_SUBNET_ID="ocid1.subnet.oc1.iad.aaaaaaaapntwxrvyrnawcmhviwxwmf2vjdsigg32577rkqwrjzihypehgsta"
export MNT_TARGET_COMPARTMENT_ID="ocid1.compartment.oc1..aaaaaaaayt3yoxoggojnw5ym3qzd3ayzsbrugdgrjwovkphovctgmyk7vb4q"
export LUSTRE_VOLUME_HANDLE=""
export LUSTRE_VOLUME_HANDLE_ARM=""
export LUSTRE_SUBNET_CIDR=""

export STATIC_SNAPSHOT_COMPARTMENT_ID=

# Workload Identity Principal Feature only available for ENHANCED_CLUSTER
export CLUSTER_TYPE="ENHANCED_CLUSTER"

# For SKE node, node_info, node_lifecycle controller tests against PDE
# To setup PDE and point your localhost:25000 to the PDE CP API refer: Refer: https://bitbucket.oci.oraclecorp.com/projects/OKE/repos/oke-control-plane/browse/personal-environments/README.md
# export CE_ENDPOINT_OVERRIDE="http://localhost:25000"
