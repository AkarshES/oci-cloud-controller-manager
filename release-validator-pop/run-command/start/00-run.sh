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

export OCI_CLI_USE_INSTANCE_METADATA_SERVICE=true

if [ -n "$cpo_image_1" ]; then
  all_images=()

  for var in $(compgen -v cpo_image_); do
    all_images+=("${!var}")
  done

  repo_name="oke-public-cloud-provider-oci"

  echo "Querying OCIR"
  existing_tags=$(oci artifacts container image list \
    --compartment-id "$COMPARTMENT_OCID" \
    --region "$REGION" \
    --repository-name "$repo_name" \
    --all \
    --auth instance_principal \
    --query 'data.items[*].[["version"], ["digest"]]' \
    --output json)

  echo "Creating Map"
  declare -A existing_tags_map
  while IFS= read -r item; do
    image_tag=$(jq -r '.[0][0]' <<< "$item")
    digest=$(jq -r '.[1][0]' <<< "$item")
    existing_tags_map["$image_tag-$digest"]=true
  done < <(jq -c '.[]' <<< "$existing_tags")

  missing_tags=()

  echo "Processing tags"
  for tag in "${all_images[@]}"; do
    image_tag=${tag%%@*}
    digest=${tag#*@}
    if [[ ! ${existing_tags_map["$image_tag-$digest"]} ]]; then
      missing_tags+=("$image_tag")
    fi
  done

  missing_tags_with_error=()

  for tag in "${missing_tags[@]}"; do
    if [[ $tag =~ oke-multiarch ]]; then
      echo "Warning: Missing image: $tag"
    elif [[ $tag =~ ^v([0-9]+)\.([0-9]+)- ]]; then
      major_version=${BASH_REMATCH[1]}
      minor_version=${BASH_REMATCH[2]}
      if (( $major_version < 1 || ($major_version == 1 && $minor_version < 28) )); then
        echo "Warning: Missing image: $tag"
      else
        echo "Error: Missing image: $tag"
        missing_tags_with_error+=("$tag")
      fi
    else
      echo "Error: Unknown tag format: $tag"
      missing_tags_with_error+=("$tag")
    fi
  done

  if (( ${#missing_tags_with_error[@]} > 0 )); then
    exit 1
  fi

  if (( ${#missing_tags_with_error[@]} > 0 )); then
    exit 1
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