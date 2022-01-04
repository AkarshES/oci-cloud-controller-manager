#!/bin/bash
function check-env () {
    if [ -z "$2" ]; then
        echo "$1 must be defined"
        exit 1
    fi
}

# ***_K8S_VERSION_INDEX*** ENV variable is the index of k8s version in version list(e.g. ["1.10.11", "1.11.8"])
# 0 means the first version, 1 means the 2nd version, -1 means the latest version
# Here we would check the index is set or not. If not, would use -1
function check-env-k8s-version-index-exist () {
    if [ -z $OKE_CLUSTER_K8S_VERSION_INDEX ]; then
        export OKE_CLUSTER_K8S_VERSION_INDEX=-1 
        echo "OKE_CLUSTER_K8S_VERSION_INDEX is not set. set -1 by default"       
    fi     

    if [ -z $OKE_NODEPOOL_K8S_VERSION_INDEX ]; then
        export OKE_NODEPOOL_K8S_VERSION_INDEX=-1        
        echo "OKE_NODEPOOL_K8S_VERSION_INDEX is not set. set -1 by default"       
    fi
}

# install the packages needed by run.sh and oci cli
function install_dependencies () {

    # this needs to be set for oci cli to work with python3
    export LC_ALL=en_US.UTF-8

    # python3 is needed by OCI CLI
    echo "Install python3 from yum"
    yum install -y python3

    # openssl is needed to convert the oci key to pem format
    echo "Install openssl from yum"
    yum install -y openssl

}

# install the oci cli - needed by kubeconfig v2
function install_oci_cli () {
    # consider where to save the install.sh, and where to install the oci binary
    mkdir oci-cli-install
    curl -L "https://raw.githubusercontent.com/oracle/oci-cli/master/scripts/install/install.sh" > oci-cli-install/install.sh
    chmod u+x ./oci-cli-install/install.sh
    ./oci-cli-install/install.sh --install-dir /usr/local/oci --accept-all-defaults
}

# create the oci config file for authenticating the cli calls
function createOCIConfig() {
    # OCI_CONFIG_DIR="$HOME/e2e/oci"
    OCI_CONFIG_DIR="$HOME/.oci"

    # Create config directory.
    mkdir -p ${OCI_CONFIG_DIR}
    if [ $? -ne 0 ]; then
         echo "Could not create oci config directory at ${OCI_CONFIG_DIR}"
         exit 1
    fi
    echo "Created OCI config directory at ${OCI_CONFIG_DIR}"

    # Create OCI key (PEM) file.
    KEY_PEM_FILE=${OCI_CONFIG_DIR}/oci_api_key.pem

    echo $OCI_KEY | sed 's/ //g' | openssl enc -base64 -d -A > $KEY_PEM_FILE
    echo "Created oci key file at $KEY_PEM_FILE"
    oci setup repair-file-permissions --file ${KEY_PEM_FILE}

    # Create OCI config file.
    CONFIG_FILE=${OCI_CONFIG_DIR}/config
    CONFIG_CONTENT="[DEFAULT]\nuser=$OCI_USER\nfingerprint=$OCI_FINGERPRINT\nkey_file=$KEY_PEM_FILE\ntenancy=$OCI_TENANCY\nregion=$OCI_REGION\n"
    echo -e $CONFIG_CONTENT > $CONFIG_FILE
    echo "Created oci config file at $CONFIG_FILE"
    export export CONFIG_FILE
    oci setup repair-file-permissions --file ${CONFIG_FILE}

}

# test that the cli can authenticate
function test_oci () {
    echo "testing oci cli"
    echo oci ce cluster list --compartment-id ocid1.compartment.oc1..aaaaaaaar5p4lkcp2tvva547lmorv6mb7e67iwy3z3mmb3lml73jtwi6quvq
    oci ce cluster list --compartment-id ocid1.compartment.oc1..aaaaaaaar5p4lkcp2tvva547lmorv6mb7e67iwy3z3mmb3lml73jtwi6quvq
}

function check_environment () {
    ENABLE_CREATE_CLUSTER=${ENABLE_CREATE_CLUSTER:-"true"}
    if [ "$ENABLE_CREATE_CLUSTER" == "true" ]; then
        check-env "OCI_USER"            $OCI_USER
        check-env "OCI_FINGERPRINT"     $OCI_FINGERPRINT
        check-env "OCI_KEY"             $OCI_KEY
        check-env "OCI_TENANCY"         $OCI_TENANCY
        check-env "OCI_REGION"          $OCI_REGION
        check-env "ADLOCATION"          $ADLOCATION
        check-env "COMPARTMENT1"        $COMPARTMENT
        check-env "LBSUBNET1"           $LBSUBNET1
        check-env "LBSUBNET2"           $LBSUBNET2
        check-env "OCI_SUBNET1"         $OCI_SUBNET1
        check-env "OCI_SUBNET2"         $OCI_SUBNET2
        check-env "OCI_SUBNET3"         $OCI_SUBNET3
        check-env "OCI_K8SSUBNET"       $OCI_K8SSUBNET
        check-env "PUB_SSHKEY"          $PUB_SSHKEY
        check-env "SECRETS_LOCAL"       $SECRETS_LOCAL
        check-env "REGION_SECRETS"      $REGION_SECRETS
        check-env "DELEGATION_GROUP_ID" $DELEGATION_GROUP_ID
        check-env "FSS_VOLUME_HANDLE"   $FSS_VOLUME_HANDLE
        check-env-k8s-version-index-exist
        if [ -z "$CLUSTER_KUBECONFIG" ]; then
            CLUSTER_KUBECONFIG="/tmp/clusterkubeconfig"
        fi
        if [ -z "$CLOUD_CONFIG" ]; then
            CLOUD_CONFIG="/tmp/cloudconfig"
        fi
    else
        check-env "CLUSTER_KUBECONFIG"    $CLUSTER_KUBECONFIG
        check-env "CLOUD_CONFIG"          $CLOUD_CONFIG
        check-env "ADLOCATION"            $ADLOCATION
    fi
}

function set_image_pull_repo_and_delete_namespace_flag () {
    if [ -z "$IMAGE_PULL_REPO" ]; then
        IMAGE_PULL_REPO=""
    fi

    DELETE_NAMESPACE=${DELETE_NAMESPACE:-"true"}
}

function run_e2e_tests() {
    export OCI_KEY_FILE=$(mktemp /tmp/ocikey.XXXXXXXXXX) || { echo "Failed to create temp file"; exit 1; }
    echo $OCI_KEY | sed 's/ //g' | openssl enc -base64 -d -A >> $OCI_KEY_FILE

    # These environment variables are used by the oci-go-sdk lib
    # For more information, you can look at the file:
    # vendor/github.com/oracle/oci-go-sdk/common/client.go
    # func DefaultConfigProvider()
    export TF_VAR_user_ocid="${OCI_USER}"
    export TF_VAR_fingerprint="${OCI_FINGERPRINT}"
    export TF_VAR_private_key_path="${OCI_KEY_FILE}"
    export TF_VAR_tenancy_ocid="${OCI_TENANCY}"
    export TF_VAR_compartment_ocid="${COMPARTMENT}"
    export TF_VAR_region="${OCI_REGION}"
    export TF_VAR_ssh_public_key="${PUB_SSHKEY}"

	if [[ -z "${E2E_NODE_COUNT}" ]]; then
	  E2E_NODE_COUNT=1
	fi

	echo "Starting tests with node count: ${E2E_NODE_COUNT}"
    ginkgo -v -progress --trace -nodes=${E2E_NODE_COUNT} "${FOCUS_OPT}" "${FOCUS_FP_OPT}"  \
        test/e2e/cloud-provider-oci -- \
        --okeendpoint=${OKE_ENDPOINT} \
        --ociuser=${OCI_USER} \
        --ocifingerprint=${OCI_FINGERPRINT} \
        --ocikeyfile=${OCI_KEY_FILE} \
        --ocikeypassphrase=${OCI_KEY_PASSPHRASE} \
        --ocitenancy=${OCI_TENANCY} \
        --ociregion=${OCI_REGION} \
        --compartment1=${COMPARTMENT} \
        --vcn=${VCN} \
        --lbsubnet1=${LBSUBNET1} \
        --lbsubnet2=${LBSUBNET2} \
        --lbrgnsubnet=${LBRGNSUBNET} \
        --nodeshape=${NODE_SHAPE} \
        --subnet1=${OCI_SUBNET1} \
        --subnet2=${OCI_SUBNET2} \
        --subnet3=${OCI_SUBNET3} \
        --k8ssubnet=${OCI_K8SSUBNET} \
        --nodesubnet=${OCI_NODESUBNET} \
        --okeClusterK8sVersionIndex=${OKE_CLUSTER_K8S_VERSION_INDEX} \
        --okeNodePoolK8sVersionIndex=${OKE_NODEPOOL_K8S_VERSION_INDEX} \
        --pubsshkey="${PUB_SSHKEY}" \
        --secrets-dir=${SECRETS_LOCAL} \
        --kubeconfig_file="${SECRETS_LOCAL}/k8-infra/${REGION_SECRETS}/kubeconfig.TNL" \
        --delegate-group-id=${DELEGATION_GROUP_ID} \
        --enable-create-cluster=${ENABLE_CREATE_CLUSTER} \
        --adlocation=${ADLOCATION} \
        --cluster-kubeconfig=${CLUSTER_KUBECONFIG} \
        --cloud-config=${CLOUD_CONFIG} \
        --delete-namespace=${DELETE_NAMESPACE} \
        --image-pull-repo=${IMAGE_PULL_REPO} \
        --cmek-kms-key=${CMEK_KMS_KEY} \
        --nsg-ocids=${NSG_OCIDS} \
        --reserved-ip=${RESERVED_IP} \
        --architecture=${ARCHITECTURE} \
        --volume-handle=${FSS_VOLUME_HANDLE}
    retval=$?
    rm -f $OCI_KEY_FILE
    return $retval
}

function run_e2e_tests_existing_cluster() {
    ginkgo -v -progress --trace "${FOCUS_OPT}" "${FOCUS_FP_OPT}"  \
        test/e2e/cloud-provider-oci -- \
        --enable-create-cluster=${ENABLE_CREATE_CLUSTER} \
        --cluster-kubeconfig=${CLUSTER_KUBECONFIG} \
        --cloud-config=${CLOUD_CONFIG} \
        --adlocation=${ADLOCATION} \
        --delete-namespace=${DELETE_NAMESPACE} \
        --image-pull-repo=${IMAGE_PULL_REPO} \
        --cmek-kms-key=${CMEK_KMS_KEY} \
        --nsg-ocids=${NSG_OCIDS} \
        --reserved-ip=${RESERVED_IP} \
        --architecture=${ARCHITECTURE} \
        --volume-handle=${FSS_VOLUME_HANDLE}
    retval=$?
    return $retval
}

function setup_amd() {
    echo "Setting up for AMD Architecture in this test."
    export ARCHITECTURE=$ARCHITECTURE_AMD
    export NODE_SHAPE=$NODE_SHAPE_AMD
    if [[ "$#" -ne  "0" && "$1" == "CREATE" ]]; then
        export VCN=$VCN_AMD
        export LBRGNSUBNET=$LBRGNSUBNET_AMD
        export OCI_NODESUBNET=$OCI_NODESUBNET_AMD
        export NSG_OCIDS=$NSG_OCIDS_AMD
        export OKE_ENDPOINT=$OKE_ENDPOINT_AMD
        declare_setup "CREATE"
    elif [[ "$#" -ne  "0" && "$1" == "EXIST" ]]; then
        export CLUSTER_KUBECONFIG=$CLUSTER_KUBECONFIG_AMD
        export CLOUD_CONFIG=$CLOUD_CONFIG_AMD
        declare_setup
    fi
}

function setup_arm() {
    echo "Setting up for ARM Architecture in this test."
    export ARCHITECTURE=$ARCHITECTURE_ARM
    export NODE_SHAPE=$NODE_SHAPE_ARM
    if [[ "$#" -ne  "0" && "$1" == "CREATE" ]]; then
        export VCN=$VCN_ARM
        export LBRGNSUBNET=$LBRGNSUBNET_ARM
        export OCI_NODESUBNET=$OCI_NODESUBNET_ARM
        export NSG_OCIDS=$NSG_OCIDS_ARM
        export OKE_ENDPOINT=$OKE_ENDPOINT_ARM
        export FSS_VOLUME_HANDLE=$FSS_VOLUME_HANDLE_ARM
        declare_setup "CREATE"
    elif [[ "$#" -ne  "0" && "$1" == "EXIST" ]]; then
        export CLUSTER_KUBECONFIG=$CLUSTER_KUBECONFIG_ARM
        export CLOUD_CONFIG=$CLOUD_CONFIG_ARM
        declare_setup
    fi
}

function declare_setup () {
    check-env "ARCHITECTURE"            $ARCHITECTURE
    if [[ "$#" -ne  "0" && "$1" == "CREATE" ]]; then
        check-env "VCN"                     $VCN
        check-env "LBRGNSUBNET"             $LBRGNSUBNET
        check-env "OCI_NODESUBNET"          $OCI_NODESUBNET
        check-env "NODE_SHAPE"              $NODE_SHAPE
        check-env "NSG_OCIDS"               $NSG_OCIDS
        check-env "OKE_ENDPOINT"            $OKE_ENDPOINT
    fi

    echo "ARCHITECTURE is ${ARCHITECTURE}"
    echo "OKE_ENDPOINT is ${OKE_ENDPOINT}"
    echo "VCN is ${VCN}"
    echo "LBRGNSUBNET is ${LBRGNSUBNET}"
    echo "OCI_NODESUBNET is ${OCI_NODESUBNET}"
    echo "NODE_SHAPE is ${NODE_SHAPE}"
    echo "NSG_OCIDS is ${NSG_OCIDS}"
}

function set_focus () {
    # The FOCUS environment variable can be set with a regex to tun selected tests
    # e.g. export FOCUS="\[cloudprovider\]"
    export FOCUS_OPT=""
    export FOCUS_FP_OPT=""
    if [ ! -z "${FOCUS}" ]; then
        # Because we tag our test descriptions with tags that are surrounded
        # by square brackets, we have to escape the brackets when we set the
        # FOCUS variable to match on a bracket rather than have it interpreted
        # as a regex character class. The check below looks to see if the FOCUS
        # has square brackets which aren't yet escaped and fixes them if needed.
        re1='^\[.+\]$' # [ccm]
        if [[ "${FOCUS}" =~ $re1 ]]; then
            echo -E "Escaping square brackes in ${FOCUS} to work as a regex match."
            FOCUS=$(echo $FOCUS|sed -e 's/\[/\\[/g' -e 's/\]/\\]/g')
            echo -E "Modified FOCUS value to: ${FOCUS}"
        fi

        echo "Running focused tests: ${FOCUS}"
        FOCUS_OPT="-focus=${FOCUS}"

        # The FILES environment variable can be defined to interpret the regex as a
        # set of files.
        # e.g. export FILES="true"
        if [[ ! -z "${FILES}" && "${FILES}" == "true" ]]; then
            echo "Running focused test regex as filepath expression."
            FOCUS_FP_OPT="-regexScansFilePath=${FILES}"
        fi
    fi
}

function declare_environment () {
    if [ "$ENABLE_CREATE_CLUSTER" == "true" ]; then
        echo "TMP_DEP_DIR is ${TMP_DEP_DIR}"
        echo "OCI_USER is ${OCI_USER}"
        echo "OCI_FINGERPRINT is ${OCI_FINGERPRINT}"
        echo "OCI_KEY is   ${OCI_KEY}"
        echo "OCI_TENANCY is ${OCI_TENANCY}"
        echo "OCI_REGION  is  ${OCI_REGION}"
        echo "COMPARTMENT is ${COMPARTMENT}"
        echo "LBSUBNET1  is  ${LBSUBNET1}"
        echo "LBSUBNET2  is  ${LBSUBNET2}"
        echo "OCI_SUBNET1  is ${OCI_SUBNET1}"
        echo "OCI_SUBNET2  is  ${OCI_SUBNET2}"
        echo "OCI_SUBNET3 is  ${OCI_SUBNET3}"
        echo "OCI_K8SSUBNET is ${OCI_K8SSUBNET}"
        echo "OKE_CLUSTER_K8S_VERSION_INDEX is ${OKE_CLUSTER_K8S_VERSION_INDEX}"
        echo "OKE_NODEPOOL_K8S_VERSION_INDEX is ${OKE_NODEPOOL_K8S_VERSION_INDEX}"
        echo "PUB_SSHKEY is  ${PUB_SSHKEY}"
        echo "SECRETS_LOCAL is  ${SECRETS_LOCAL}"
        echo "REGION_SECRETS is ${REGION_SECRETS}"
        echo "DELEGATION_GROUP_ID is ${DELEGATION_GROUP_ID}"
    else
        echo "CLUSTER_KUBECONFIG is ${CLUSTER_KUBECONFIG}"
        echo "CLOUD_CONFIG is ${CLOUD_CONFIG}"
    fi

    if [[ $LOCAL_RUN != 1 ]]; then
        if [[ ! -z $TC_BUILD ]]; then
        # kubeconfig v2 requres oci cli and an oci config
        # The docker image installs these already
        install_dependencies
        install_oci_cli
        createOCIConfig
        fi

    # uncomment this to verify authentication if needed
    # test_oci
    fi
}

function run_tests () {
    set_image_pull_repo_and_delete_namespace_flag
    set_focus
    if [[ "$ENABLE_CREATE_CLUSTER" == "true" ]]; then
        # run the ginko test framework with cluster create
        # run AMD tests
        if [[ "$SCOPE" == "BOTH" || "$SCOPE" == "AMD" ]]; then
            setup_amd "CREATE"
            check_environment
            declare_environment
            run_e2e_tests
            retval_amd=$?
        fi
        # run ARM tests
        if [[ "$SCOPE" == "BOTH" || "$SCOPE" == "ARM" ]]; then
            setup_arm "CREATE"
            check_environment
            declare_environment
            run_e2e_tests
            retval_arm=$?
        fi
    else
        # run the ginko test framework for existing cluster
        # run ARM tests
        if [[ "$SCOPE" == "BOTH" || "$SCOPE" == "ARM" ]]; then
            setup_arm "EXIST"
            check_environment
            declare_environment
            run_e2e_tests_existing_cluster
            retval_arm=$?
        fi
        # run AMD tests
        if [[ "$SCOPE" == "BOTH" || "$SCOPE" == "AMD" ]]; then
            setup_amd "EXIST"
            check_environment
            declare_environment
            run_e2e_tests_existing_cluster
            retval_amd=$?
        fi
    fi

    RED='\033[0;31m'
    NC='\033[0m' # No Color
    if [[ "$SCOPE" == "BOTH" ]]; then
        if [[ $retval_amd == 0 && $retval_arm == 0 ]]; then
            printf "ARM and AMD tests are Successful!"
            return $retval_amd
        fi

        if [[ $retval_amd != 0 ]]; then
            printf "${RED}AMD Failed${NC}"
            return $retval_amd
        fi

        if [[ $retval_arm != 0 ]]; then
            printf "${RED}ARM Failed${NC}"
            return $retval_arm
        fi
    fi

    if [[ "$SCOPE" == "ARM" ]]; then
        if [[ $retval_arm != 0 ]]; then
            printf "${RED}ARM Failed${NC}"
            return $retval_arm
        else
            echo "ARM tests are Successful"
            return $retval_arm
        fi
    fi

    if [[ "$SCOPE" == "AMD" ]]; then
        if [[ $retval_amd != 0 ]]; then
            printf "${RED}AMD Failed${NC}"
            return $retval_amd
        else
            echo "AMD tests are Successful"
            return $retval_amd
        fi
    fi
}

run_tests
