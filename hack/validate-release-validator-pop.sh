#!/bin/bash
echo "Validating release validator POP version"
set -euo pipefail

dnf install -y jq
# Function to compare image_list JSON and manifest.csv
compare_manifest_json() {
  local manifest="$1"
  local json="$2"

  # sanity checks
  [ -f "$manifest" ] || { >&2 echo "Manifest file not found: $manifest"; return 1; }
  [ -f "$json" ]     || { >&2 echo "Image JSON file not found: $json"; return 1; }

  # count entries
  local manifest_count json_count ret=0
  manifest_count=$(grep -cve '^[[:space:]]*$' "$manifest")
  json_count=$(jq '.images | length' "$json")
  if [ "$manifest_count" -ne "$json_count" ]; then
    >&2 echo "Count mismatch: MANIFEST.csv has $manifest_count entries; JSON has $json_count entries"
    ret=1
  fi

  # collect missing entries
  local -a missing=()
  while IFS=, read -r raw_name raw_expected; do
    name=$(echo "$raw_name"     | xargs)
    expected=$(echo "$raw_expected" | xargs)
    if ! grep -qE "^${name}[[:space:]]*,[[:space:]]*${expected}$" "$manifest"; then
      missing+=("$name → $expected")
    fi
  done < <(jq -r '.images[] | to_entries[] | "\(.key),\(.value)"' "$json")

  # report missing entries if any
  if [ "${#missing[@]}" -gt 0 ]; then
    >&2 echo "ERROR: Missing entries in manifest.csv:"
    for entry in "${missing[@]}"; do
      >&2 echo "   - $entry"
    done
    ret=1
  fi

  return $ret
}


# Paths
manifest="MANIFEST.csv"
tf_file="./shepherd/limits/shared_modules/properties_values/default_values.tf"

# Validate required files
for f in "$manifest" "$tf_file"; do
  [[ -f $f ]] || { >&2 echo "File not found: $f"; exit 1; }
done


# Extract pop_version from Terraform defaults
pop_version=$(awk -F '"' '/pop_version/ { gsub(/^[ \t]+|[ \t]+$/, "", $2); print $2; exit }' "$tf_file")
if [[ -z "$pop_version" ]]; then
  >&2 echo "pop_version not found in $tf_file"
  exit 1
fi
echo "Using pop_version: $pop_version"

# Define archive based on pop_version
archive="release-validator-ccm-csi-${pop_version}.tar.gz"
url="https://artifactory-builds.oci.oraclecorp.com:443/odo-artifacts-signed-generic-local/$archive"

# Prepare temp working directory
temp_dir="./temp-working"
echo "Preparing temp directory: $temp_dir"
rm -rf "$temp_dir"
mkdir -p "$temp_dir"

# Download into temp and extract there
download_path="$temp_dir/$archive"
echo "Downloading $url → $download_path..."
curl -fSL "$url" -o "$download_path"

echo "Extracting $download_path into $temp_dir..."
tar -xzf "$download_path" -C "$temp_dir"

# Compare manifest vs extracted JSON
json_file="$temp_dir/image_versions.json"
echo "Verifying images_version.json from POP artifact against MANIFEST.csv..."
if ! compare_manifest_json "$manifest" "$json_file"; then
  >&2 echo "JSON vs manifest.csv comparison failed. Aborting."
  rm -rf "$temp_dir"
  exit 1
fi

# Cleanup downloaded archive
echo "Validation for release validator POP version is successful!"
rm -rf "$temp_dir"
