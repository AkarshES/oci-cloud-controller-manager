import oci
import base64
import sys

secret_id = "ocid1.vaultsecret.oc1.phx.amaaaaaa7aeocuiazbqfbqm3nttrh6yx252dwkxlq7lhahblpd52nnqqopta"

# By default this will hit the auth service in the region the instance is running.
signer = oci.auth.signers.InstancePrincipalsSecurityTokenSigner()

# Get instance principal context
secret_client = oci.secrets.SecretsClient(config={}, signer=signer)

# Retrieve secret
def read_secret_value(secret_client, secret_id):
    response = secret_client.get_secret_bundle(secret_id)
    base64_Secret_content = response.data.secret_bundle_content.content
    base64_secret_bytes = base64_Secret_content.encode('ascii')
    base64_message_bytes = base64.b64decode(base64_secret_bytes)
    secret_content = base64_message_bytes.decode('ascii')
    return secret_content

secret_contents = read_secret_value(secret_client, secret_id)

bb_access_key_file = open("bb_access_key", "w")
bb_access_key_file.write(format(secret_contents) + "\n")
bb_access_key_file.close()
