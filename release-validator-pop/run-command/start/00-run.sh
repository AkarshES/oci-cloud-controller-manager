#!/bin/bash

set -e
set -o pipefail

#exec &> >(tee -a "${ODO_APPLICATION_ROOT}/var/start.log")

echo "Starting release validation"

if [[ -z "$ODO_APPLICATION_ROOT" ]]; then
  echo "No ODO_APPLICATION_ROOT defined, cannot continue"
  exit 1
fi

JSON_FILE="${ODO_APPLICATION_ROOT}/image_versions.json"

COMPARTMENT_OCID=$STEWARD_TENANCY_OCID

export OCI_CLI_USE_INSTANCE_METADATA_SERVICE=true
export REQUESTS_CA_BUNDLE="/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"

if [ -n "$cpo_image_1" ]; then
  all_images=()

  for var in $(compgen -v cpo_image_); do
    all_images+=("${!var}")
  done

  repo_name="oke-public-cloud-provider-oci"

  existing_tags=$(oci artifacts container image list \
    --compartment-id "$COMPARTMENT_OCID" \
    --region "$REGION" \
    --repository-name "$repo_name" \
    --all \
    --auth instance_principal \
    --query 'data.items[*].[["version"], ["digest"]]' \
    --output json)

  missing_tags=()

  for tag in "${all_images[@]}"; do
    image_tag=${tag%%@*}
    digest=${tag#*@}

    found=false
    for item in $(jq -c '.[]' <<< "$existing_tags"); do
      if [[ $(jq -r '.[0][0]' <<< "$item") == "$image_tag" && $(jq -r '.[1][0]' <<< "$item") == "$digest" ]]; then
        found=true
        break
      fi
    done
    if ! $found; then
      missing_tags+=("$tag")
    fi
  done

  if [ ${#missing_tags[@]} -gt 0 ]; then
    echo "The following images are missing from OCIR:"
    printf '%s\n' "${missing_tags[@]}"
    exit 1
  else
    echo "All images are present in OCIR."
  fi
else
  fetch_repository_tags() {
    local repo_name=$1
    oci artifacts container image list \
        --compartment-id "$COMPARTMENT_OCID" \
        --region "$REGION" \
        --repository-name "$repo_name" \
        --all \
        --auth instance_principal \
        --query 'data.items[*]."display-name"' \
        --output json | jq -r '.[]' | awk -F':' '{print $2}'
  }

  jq -r '.images[] | keys[]' "$JSON_FILE" | sort -u | while read -r repo_name; do
    echo "Checking repository: $repo_name"

    repo_tags=$(fetch_repository_tags "$repo_name")

    expected_tags=$(jq -r --arg repo "$repo_name" '.images[][$repo] // empty' "$JSON_FILE")

    for tag in $expected_tags; do
      if echo "$repo_tags" | grep -q "^$tag$"; then
        continue
      else
        echo "  The following image is not present in OCIR: $tag"
        exit 1
      fi
    done
    echo
  done

  echo "All images found in OCIR."
fi