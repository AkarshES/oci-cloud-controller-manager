# create the oci config file for authenticating the cli calls
function createOCIConfig() {
    # OCI_CONFIG_DIR="$HOME/e2e/oci"
    OCI_CONFIG_DIR="$HOME/.oci"
    ROOT="/root"
    echo "Current user: $(whoami); HOME Loc: ${HOME}"

    # Create config directory.
    mkdir -p ${OCI_CONFIG_DIR} "$ROOT/.oci"
    if [ $? -ne 0 ]; then
         echo "[OCI_CONFIG] Could not create oci config directory at ${OCI_CONFIG_DIR}"
         exit 1
    fi
    echo "[OCI_CONFIG] Created OCI config directory at ${OCI_CONFIG_DIR} and ${ROOT}/.oci"

    # Create OCI key (PEM) file.
    KEY_PEM_FILE=${OCI_CONFIG_DIR}/oci_api_key.pem
    ROOT_PEM_FILE="$ROOT/.oci"/oci_api_key.pem

    echo $OCI_KEY | base64 -d > $KEY_PEM_FILE
    cp "$KEY_PEM_FILE" "$ROOT_PEM_FILE"
    echo "[OCI_CONFIG] Created oci key file at $KEY_PEM_FILE and $ROOT_PEM_FILE"
    oci setup repair-file-permissions --file "${KEY_PEM_FILE}"
    oci setup repair-file-permissions --file "${ROOT_PEM_FILE}"
    echo "Key file: $(cat "$ROOT_PEM_FILE")"

    # Create OCI config file.
    CONFIG_FILE=${OCI_CONFIG_DIR}/config
    ROOT_CONFIG_FILE="$ROOT/.oci"/config
    CONFIG_CONTENT="[DEFAULT]\nuser=$OCI_USER\nfingerprint=$OCI_FINGERPRINT\nkey_file=$KEY_PEM_FILE\ntenancy=$OCI_TENANCY\nregion=$OCI_REGION\n"
    echo -e "$CONFIG_CONTENT" > "$CONFIG_FILE"
    echo -e "$CONFIG_CONTENT" > "$ROOT_CONFIG_FILE"
    echo "Created oci config file at $CONFIG_FILE and $ROOT_CONFIG_FILE"
    echo "Config file: $(cat "$ROOT_CONFIG_FILE")"
    oci setup repair-file-permissions --file "${CONFIG_FILE}"
    oci setup repair-file-permissions --file "${ROOT_CONFIG_FILE}"

}

# test that the cli can authenticate
function test_oci () {
    echo "testing oci cli"
    set -x
    oci ce cluster list -c "${COMPARTMENT}" --endpoint https://containerengine-integ.us-phoenix-1.oci.oraclecloud.com --debug
    set +x
}

createOCIConfig
test_oci
