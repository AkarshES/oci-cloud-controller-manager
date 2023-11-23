#!/bin/bash
# ODO will attempt to rerun this script if it fails; we don't want that as it will take the max time (60 minutes)
# to get a failure reported through pipelines. This isn't configurable, so instead we'll write the exit code to a file
# and check for that file on startup to return the same error code
if [ -f exit_status ]; then
  exit_status=$(cat exit_status)
  echo "Job finished. Download the logs by running 'oci os object get -ns odx-mockcustomer -bn oke-dp-e2e-logs --name ${ODO_APPLICATION_CHANGE}_ccm_e2e_logs.txt --file ./ccm-e2e-logs.txt'"
  exit $exit_status
fi

run_tests() {
  # ODO has a limit on the length of an environment variable value. So these must be hardcoded here.
  export OCI_KEY="LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBeTk3STA4TURKbml0UFNEOU92WlFHSys0R3pIUEliV3hvUExEdEl4RUdnSmUzdmJrClJqZ3kwQWxzTUt5TURsWmpmNkQ1L1hxTFlBdllSUmlTUE5IZ2NJUzhMOVJ4ZUt5SUtQUCtEWUFteDBmTjN0NGIKbnFBR1lwbzVKZ0QzbE44VW9teENGWWhwK3RUL0JidS9RTUd1dFgyVkJBSGhOMndxamZrTFBXQVIrZFJ0eUZregovRExSNHdSK1NtZTlINDgyQ2l0dEZmL3RWeG5uakQxbkF4TTlZU0hPaU81TWgxRlBjZjllNGgxaktWSURyT2plCnlDYWdOeGl4a3pJOFZTaTZ3VVdQNk11bitzR3VBYmswVmxnaE1zOFRXODF4NUh0TmZpcjE0VVd4aVBnYzlyK3QKYTlzSTF2NURDLzVvVmVWM2dta055eXdGRlRJUndIZzQxTWdoSlFJREFRQUJBb0lCQURwSlVwK0FoTGtPRFhHSgpxZnllaVYzVmQweUhIQkltTVVlendKSXcvQk4zbEFvcENqQ3RScEhGNytHbW8xQkNFS1pmcnJseXZNQlVBcHdXCm4xQzJMSFhlN1RLN2lVbVlBa0c3S0dwUnRrU0pXZW1iRk1od05nTWcxcS92M05qNlFwc0JXbzl2ZGVWWjJpT1kKZnVDVHYxQnlQQlZ1T0NheFRTVlVDNHVibU9nVStmaXFtSjduYXJjb1BuTHVyUWcwbWVjUFk0R282elBrNWxWZApDeGlRZUJNNFhVcnByYU93WkVPcUZVaCt2YkNVbE1zM205K1lrNXpiVzk5R3VwRndzZS8weklMd045b0FNamlzCnlOV3Q0QWpYVWxNY051M2ZlaTc5YjAvWkRaWmdocXFvVGo4QmY1c1EvUTFpenI4bUxQcnNYUHg5bkdSWk5UTVkKNitXK2VrRUNnWUVBOCs3NC9rT3E3U241aEt3Sm5OenZ6VkZmOWxRb3VtNFdnZ1NET28wTHJHbjltUTBCeXNWVQpqTmFwWUZDS0tGSXhhc3lKaW5UU2U2K0IrRFBEN3h1R0dkT0hYU2VtazlEL0VQa0cxZTkyWGZuNFhzOU9wZnFQCkJvdldCNE5HemM3ZE4rNnlIMXBnM0hFU09XWGFVczBUbmtmeG9yUlp3NHMyQ2plT0FmUmFzRkVDZ1lFQTFmUjUKbm1YK1NXL3JpdGFUSTJDQ0hWSkszS0JjWmtIeGJVbDg3a0o0Mlhwb1hadWl1RnpzVXhBWDhFdjUrTHc4TjkrUApqdEFzV285ai8zRE5QeUx3RUhNRGVkaDhEeDVaNG8xVnoxaVJOZWxYOTRFYldDeklmbFJKSDFFd1NjOUE1MnVCClFMOEdrK1pib2tWbGczUExNV2dpOU1Id2xzdWp4aHoyb2tqWTRwVUNnWUVBejFXWG5jYTJTYzhibm9EN3lyQ2YKd3N3ZjNLZTcxbm9yOTJjT2czM1BKa1VRb0Q1Ri82dXFLZXRyRzhwWkk5eDQxR1gxb2hHWnRqZC9LUkFXd2UyTApGZmlOWGNpTjlhWDJwdDZEeU9NNG9MZ1BPOGJJK3ByMVpvTUU4RzNNaUxJRHBiN2s5M0ZDVXh4b2VSRHhlTitHCmhVcHhQL0k3T2RIaHBOMkRmblhMZjRFQ2dZRUFoaUMrUWFQVkZ6cTNvaEZFcXEvc3dlNytDWFBxbVJ5TCtxUlEKVXBtUFkxOU9vQ2hhaVZPUHY3N2VFd0crVXNYR3hvdzdWaCtCUHNDWHd1SjNlVFl3NDA4SEJkTEdhei91djQwMQpGMGlCdUJkeDB6SzM3cjRjYnIzdUhWanJlY29ZK1RzM0MxejJCYkFyRC82TFZpNDRXdC9hMGkvbTROSUcrTUxkCmh2MCtJa1VDZ1lBSS9iaGtCOEpiVmdWelZVZVNLU3B1U2JuWjkvM3JRUWhVUTVURkNkSkREbzNGclRRNC9YeVYKeUNpb1Q0aHo2TENHSWRXVnh4bDBuZ1luMFlUaTdwRWt3cGNkQjhHOGN5R0NMNzBoN3Zhb0tZZUNjSE9HUE02cApscm1naVVMbDhvemp2QlNqeDcvMWRjOWtjVnVqU29aYUpJTzB6Y2l5aHJlTEE0MnJydE1nbnc9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
  export PUB_SSHKEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC07ybO1OfwsfoVwEi95aelL7j3/Eiyzh28sqCeihN0HHtBTw8UrHSsl5lV7MerhUH7iH4hfY91vcby0K2tpoMl1FWjf26s6rDtzjawJA8Q2wSV3NUI0GP+r7VoH+YbxrfXXDhbDlYcBYjOIDhH5GL0/3vRL7kGnUmXV9RzWmfVbz/24u/uA/pK0GUvKWwXAXBYqDle6GtWBDngocSbEF0nc+SJBONH0oqPlCUxt1LpmxxOtBsA1UvxsTzrZIuP2Wf4vigVqAogqMc5g13G/flSE5dCXqLDacOyobrNhSc6GgAWF+ZnEEF2dX/aR5d08Zwh9xymmdPNFKftr4Jzb8UPh4Ha3NUZT7TJipKOz8uj2ezGYXx+ltfg55Kapx4tOP0xU7u07vu1favfd/pn7QKJWp3QYO2tKIZk5euIBR+5MZPQnAItfQbRXA1EvTRY39J7hmwBfeArs6LHKWeRjBri11UJX5wzM44orY23HeCb4hsF0QH7/Y3vQOon3q/OeCTJNJbwiVJwaw0SevbDRGCZkLxy0wLtoNdv4I+dEXF/Dbevv73gdsDbk7tCMk5I/nlV6q1oM9aMrvtwjsYysKrkCSCY3eu/d10sAfbzJzaf1htUj07OkHEgRaG3oQ3+Rn+MgEyIsOPJjeoYiHZRSZbf+F8Y/qboBKvcPQ63evbRKQ== e2e_user@oracle.com"
  export SOPS_OCI_KEY="LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBNTUwSG5WV2xDd3AycW04KzlLTVh1ZkN2Q3MwdzlLNjh2RFRZK3RFdEJacFRqN0g1CjZKcnE4NUJWNnpnVlNVL21oUFQySEdxaXpxWTlYTXJTS09BUDJuU2pJYTRPdFZLcWZLa1dEZm9GY1lHdlNMVFIKVGxBc0RNV1prbjc4cWtzYU1RaVBVZVdyeDUwK09Uc3hWUkNUbnJqQ0MwNlNvYzg4NUgyZFVONjlpS29hTS9BcQorUVpHRGNzM3ZxMGRrK1NlbFBmakdhRnZQaGhnY3EyOXJZZURqZjJtN0I2dlZUc2g0Z3VhZWZ6dzMxZ2NtTW1ICm8zN3BvbHZ1UU4xekZXZHlmSlNZaTdxWU01N1BZWW9tRzlweTZqVjExNU1yeHdpV211eDZnb3IxbS9palowSFAKanBoYjlZZ1BHVHpTK3hGaEFKTkpUNmo5b3Q3N05zUGMrTW04aHdJREFRQUJBb0lCQUhTejNPaGNCU05CREhYbwpRL0tFVXlvdnFpTnBMS2U5ZS8vaEtRdUlab1VTTTlTTnV2eUhRcWVqTElldTVKcmlGYWNjdnYybWhZNVdtVWl1Ck1hTEM4M29CbDBrYktQSVlTeXN4RDVuUFJ2cmdlMi9KSEVXM2c5cVRua1FBbEZOQ01GcFFrOWFTUVRIOWV5TGYKUG9ZOEtnV3AxcXhYMU82UE80RnVBZHlEdjBDaWI2cHhIYmhhbmRBQnpvTGFKNDVZWHZQRTRnaXE3aVN4QUh3OApRZ1J4VWVyN0czT0cvSkRUenc2WlY2dytGVHgvUXZNUnVCOUpSVWVEd3g0NlRjbDVhMjh1QXVaYVNHMFhsU0JLClpwVXU5N1lFKzlNdURMcGt2RFd1S2k4UTNkTFY1K004WkFlMWxMNlJSMmZNdUJKKzJFWWZ5cjYxNnk4eDM3UUUKN1NHbDdyRUNnWUVBL1NlZjJMMEZWRnRvQ0hJSXo0VGs4VHdiRE9tcUU3cENHcUNCRGRTN3ZRSndmWkNlYS9yWQpIV2NacFh1c25qQ1FicFFVT2Rob3NXRXVrZWVOVk5QM2g4aXR4L1UvMHhndDRvOE9KUEF5dk5GMjk2dDBNd01vCjgxeDhka3JRdW9BVTFXc25Ybll1VTQwRGtoaVZNTE5KcnBaYSt4YlhoMHVmc2lRWkRudWdGZnNDZ1lFQTZqZHQKTkpmaDh0N1B5cjdjVWhZQjZrK2JpQ3V0NzViQ2J0Wk8vMEhRdEF1bzVYVDQxK1FCTHRvTWIrYytsdW80dXNxegpFOVFaaklSbDR1cE5qMDJOZGtIR1pNTzJwTHpzZHp3VGF2YU4rMTFSQjdManltTUU2QlVCZTlQcU1pWkR3aUZYCi9HKzFsWlFWL0ZYU1liVjhLc0Y3QkJyamhyU01CdjRDaWhlOHllVUNnWUFhUVZyTnNzVHp1OHN0WFE1VzMzU1QKSkdXMTBDSW9pNS9CZlRZRlJqUDJaV05mVW5scnY0ZGNmVTNtb000RlZnb1V2ZHpmSnZlc3RlU0xrMVZRSCsvRApNR2Y5bmd6eGlzZHZnT0M1cWdQSkczeFlNWHNLczJBeTVUdXZWUkVTMXFmU2ZwdUZxNElnZmphSmwyMFpzTzZLClllT3J1UEcvZ1hOZG1XclQyclIrc3dLQmdRQ2hKWCtDbHhtRHc5K043R0drTExZbW81MHNSSGxKQmg1Q3FqcnkKRHpOc0hUV1lvakZ3UU5TN2lwVFNEWFdYMmhFc0c3aTRaTThyU2hEYjNqOTg0R3Y3T0dncS9pbFZFUk5WT2tWVgo2OHRtYmg4SlFBRFFSKzZoUzRxWXl3WXdlUGxYd1I5TGRRU25wSnEzNGNoOUo2UUZ1dGRMek1CTTl6Mzh0Wm5ECnIyOWFMUUtCZ1FEY0cvU0ZNYzhZTWh4blJPK3hQQXRkR3JORFNLblFVZllkamdsZ2ttMHBubnM2UThSOGc0dC8KWHc2UTNSWXVRTFJ6ZlNOUkZxcU1GaytVREpybVVhcjVQSEZONUo5aVVZNGpPeHRBc2R5ZUR6N1daSlNtTFNFUwpZblZKUFJKNENtbmVqVXk2dU1KR0Urckg2dlRtWks5dzhHcG95M2g0MDc0QmtoUHRDRU8rZ1E9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="

  echo "Environment"
  env

  echo "Starting OKE E2E validation scripts"
  echo "  Command: $OKE_TEST_COMMAND"

  echo "Loading image from tarball, this takes a bit"
  docker load -i ./images/oke-ccm-e2e-tests-pop-image.tar.gz
  # delete the image tarball to save some space on the odo host
  rm -rf ./images/*

  # Run a docker container, piping the logs to a file
  docker run \
          --network host \
          --mount type=bind,source=/etc/region,target=/etc/region,readonly \
          --mount type=bind,source=/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem,target=/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem,readonly \
          --rm \
          -w /go/src/github.com/oracle/oci-cloud-controller-manager \
          -e COMPARTMENT \
          -e LBRGNSUBNET \
          -e LBSUBNET1 \
          -e LBSUBNET2 \
          -e OCI_FINGERPRINT \
          -e OCI_REGION \
          -e OCI_RGNSUBNET \
          -e OCI_SUBNET1 \
          -e OCI_SUBNET2 \
          -e OCI_SUBNET3 \
          -e OCI_TENANCY \
          -e OCI_USER \
          -e OKE_CLUSTER_K8S_VERSION_INDEX \
          -e OKE_NODEPOOL_K8S_VERSION_INDEX \
          -e OKE_CANARY_DEPLOY \
          -e OKE_ENDPOINT \
          -e OKE_ENVIRONMENT \
          -e OKE_MATURITY \
          -e OKE_REGION_TOP_LEVEL \
          -e OKE_REGISTRY_PATH \
          -e OKE_TENANCY_NAME \
          -e PUB_SSHKEY="${PUB_SSHKEY}" \
          -e REGION_SECRETS \
          -e VCN \
          -e OCI_KEY \
          -e NODE_SHAPE \
          -e DELEGATION_GROUP_ID \
          -e SECRETS_LOCAL \
          -e USE_REGIONALSUBNET \
          -e FOCUS \
          -e ADLOCATION \
          -e IMAGE_PULL_REPO \
          -e KUBECONFIG \
          -e SOPS_OCI_REGION \
          -e SOPS_OCI_TENANCY \
          -e SOPS_OCI_USER \
          -e SOPS_OCI_FINGERPRINT \
          -e SOPS_OCI_KEY \
          -e RUN_PRE_UPGRADE \
          -e RUN_POST_UPGRADE \
          -e OLD_VERSION \
          -e NEW_VERSION \
          -e ODO_APPLICATION_CHANGE \
          -e KMS_SECRETS_KEY \
          -e CMEK_KMS_KEY \
          -e MNT_TARGET_ID \
          -e MNT_TARGET_SUBNET_ID \
          -e MNT_TARGET_COMPARTMENT_ID \
          -e NSG_OCIDS \
          -e RESERVED_IP \
          -e FSS_VOLUME_HANDLE \
          -e OKE_TEST_COMMAND \
          -e STATIC_SNAPSHOT_COMPARTMENT_ID \
          -e CREATE_UHP_NODEPOOL \
          -e ENABLE_PARALLEL_RUN \
          oke-ccm-e2e-tests-pop /bin/bash -c "./images/e2e-tests/e2e_pop_run.sh" 2>&1
}

cleanup() {
    exit_status=$?
    echo -n $exit_status > exit_status
    echo "Exit status was ${exit_status}" | tee -a ccm-e2e-logs.txt

    # Remove docker image
    echo "Removing oke-ccm-e2e-tests-pop image" | tee -a ccm-e2e-logs.txt
    docker rmi oke-ccm-e2e-tests-pop

    echo "Job finished. Download the logs by running 'oci os object get -ns odx-mockcustomer -bn oke-dp-e2e-logs --name ${ODO_APPLICATION_CHANGE}_ccm_e2e_logs.txt --file ./ccm-e2e-logs.txt'" | tee -a ccm-e2e-logs.txt
    exit $exit_status
}
trap cleanup EXIT

set -o pipefail
run_tests | tee -a ccm-e2e-logs.txt
