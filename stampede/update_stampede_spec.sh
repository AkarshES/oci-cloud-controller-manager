#!/bin/bash
set -euo pipefail

# Default values
manifest=""
stampede_json=""
tf_file=""

# Parse flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    -m|--manifest)
      manifest="$2"
      shift 2
      ;;
    -s|--stampede)
      stampede_json="$2"
      shift 2
      ;;
    -t|--tf)
      tf_file="$2"
      shift 2
      ;;
    *)
      >&2 echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Validate input
if [[ -z "$manifest" || -z "$stampede_json" || -z "$tf_file" ]]; then
  >&2 echo "Usage: ./gen_artifacts_tfvars.sh -m <manifest.csv> -s <stampede.json> -t <file.tf>"
  exit 1
fi

# Extract pop_version from tf file
# pop_version=$(awk -F'"' '/pop_version/ { gsub(/^[ \t]+|[ \t]+$/, "", $2); print $2; exit }' "$tf_file")
pop_version="skip"

# Build artifacts JSON array from CSV
artifacts_json=$(
awk -F, -v pop_version="$pop_version" '
  BEGIN {
    print "["
    sep = ""
  }
  /^[^#]/ {
    loc = $1
    ver = $2
    gsub(/^[ \t]+|[ \t]+$/, "", loc)
    gsub(/^[ \t]+|[ \t]+$/, "", ver)

    name_ver = ver
    gsub(/\./, "_DOT_", name_ver)

    name = loc "__" name_ver

    printf "%s{\"name\": \"%s\", \"version\": \"%s\"}", sep, name, ver
    sep = ",\n"
  }
  END {
    printf ",\n{\"name\": \"release-validator-ccm-csi\", \"version\": \"%s\"}\n", pop_version
    print "]"
  }
' "$manifest"
)

# Inject artifacts array into stampede.json using jq
tmpfile=$(mktemp)
echo "==== Adding following json to stampede specs for oke-ccm-csi flock ===="
echo "$artifacts_json" | jq
jq --argjson artifacts "$artifacts_json" \
  '(.versionSets[] | select(.projectName == "oke" and .flockName == "oke-ccm-csi" and .changeType == "Application") | .artifacts) = $artifacts' \
  "$stampede_json" > "$tmpfile"

mv "$tmpfile" "$stampede_json"
echo "✅ Updated artifacts in $stampede_json based on $manifest and $tf_file"
