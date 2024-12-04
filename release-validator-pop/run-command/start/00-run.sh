#!/bin/bash

set -e
set -o pipefail

exec &> >(tee -a "${ODO_APPLICATION_ROOT}/var/start.log")

echo "Starting release validation"

if [[ -z "$ODO_APPLICATION_ROOT" ]]; then
  echo "No ODO_APPLICATION_ROOT defined, cannot continue"
  exit 1
fi

JSON_FILE="${ODO_APPLICATION_ROOT}/image_versions.json"

COMPARTMENT_OCID=$STEWARD_TENANCY_OCID

if [ -n "$REGIONAL_IMAGE_LIST_1" ]; then
  IFS=',' read -r -a regional_image_list_1 <<< "$REGIONAL_IMAGE_LIST_1"
  IFS=',' read -r -a regional_image_list_2 <<< "$REGIONAL_IMAGE_LIST_2"

  IFS=',' read -r -a overrides_image_list_1 <<< "$OVERRIDES_IMAGE_LIST_1"
  IFS=',' read -r -a overrides_image_list_2 <<< "$OVERRIDES_IMAGE_LIST_2"

  unset IFS

  all_image_tags=()

  for array in regional_image_list_1 regional_image_list_2 overrides_image_list_1 overrides_image_list_2; do
    eval "all_image_tags+=(\"\${$array[@]}\")"
  done

  repo_name="oke-public-cloud-provider-oci"

  existing_tags=$(oci artifacts container image list \
    --compartment-id "$COMPARTMENT_OCID" \
    --region "$REGION" \
    --repository-name "$repo_name" \
    --all \
    --auth instance_principal \
    --query 'data.items[*]."display-name"' \
    --output json | jq -r '.[]' | awk -F':' '{print $2}')

  missing_tags=()

  while IFS= read -r tag; do
    found=false
    if grep -q "^${tag}$" <<< "$existing_tags"; then
      found=true
    fi
    if ! $found; then
      missing_tags+=("$tag")
    fi
  done < <(printf "%s\n" "${all_image_tags[@]}")

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