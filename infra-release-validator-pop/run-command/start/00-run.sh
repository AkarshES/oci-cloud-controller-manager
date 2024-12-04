#!/bin/bash

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
COMPARTMENT_OCID=${STEWARD_TENANCY_OCID}

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
