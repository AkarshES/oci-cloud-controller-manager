#!/bin/bash
set -e
if [[ "$#" -ne 1 ]]; then
    >&2 echo "Usage: ./gen_images_tfvars.sh <path_to_wheelbarrow_manifest>"
    exit
fi
manifest=$1

echo "{\"images\": ["

awk ' \
  BEGIN { SEP=""; FS=",";} \
  /.\/*/{ \
    gsub(/^[ \t]+|[ \t]+$/, "", $1); \
    gsub(/^[ \t]+|[ \t]+$/, "", $2); \
    gsub("\\\.", "_DOT_", $2); \
    printf "%s  {\"name\": \"%s__%s\", \"location\": \"%s\"}", SEP, $1, $2, $1; \
    SEP=",\n"; \
  } \
  END {printf "\n"}' ${manifest}

echo "]}"

input_csv="$1"
output_json="release-validator-pop/image_versions.json"

if [ ! -f "$input_csv" ] || [ ! -r "$input_csv" ]; then
  echo "Error: Input file does not exist or cannot be read."
  exit 1
fi

awk -F, '{gsub(/^ *| *$/,"",$1); gsub(/^ *| *$/,"",$2); print "{\""$1"\": \""$2"\"}"}' "$input_csv" | jq -s '. | {images: .}' > "$output_json"

echo "JSON file created: $output_json"