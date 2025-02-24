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

json_images=($(jq -r '.images[] | to_entries[] | .key + "," + .value' $output_json))

cat $input_csv | while IFS=, read -r image_name image_version; do
  image_name=$(echo "$image_name" | tr -d '[:space:]')
  image_version=$(echo "$image_version" | tr -d '[:space:]')

  found=false
  for ((i=0; i<${#json_images[@]}; i++)); do
    parts=(${json_images[i]//,/ })
    if [ "${parts[0]}" == "$image_name" ] && [ "${parts[1]}" == "$image_version" ]; then
      found=true
      break
    fi
  done

  if ! $found; then
      echo "Following image is missing in image_versions.json: $image_name - $image_version"
    fi
done