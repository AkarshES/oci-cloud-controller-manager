#!/bin/bash

set -e
set -o pipefail

exec &> >(tee -a "${ODO_APPLICATION_ROOT}/var/start.log")

echo "Starting image release validation"

if [[ -z "$ODO_APPLICATION_ROOT" ]]; then
  echo "No ODO_APPLICATION_ROOT defined, cannot continue"
  exit 1
fi

JSON_FILE="${ODO_APPLICATION_ROOT}/image_versions.json"

COMPARTMENT_OCID=$STEWARD_TENANCY_OCID

fetch_repository_tags() {
    local repo_name=$1
    oci artifacts container image list \
        --compartment-id "$COMPARTMENT_OCID" \
        --region "$REGION" \
        --repository-name "$repo_name" \
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
            echo "  Tag missing: $tag"
            exit 1
        fi
    done
    echo
done

echo "All images found in OCIR."