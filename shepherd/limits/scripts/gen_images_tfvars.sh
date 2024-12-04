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


if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <path-to-wheelbarrow-manifest>"
  exit 1
fi

input_csv="$1"
output_json="release-validator-pop/image_versions.json"

echo '{"images": [' > "$output_json"

# Read the CSV file line by line and store each line as a JSON object in an array
json_objects=()
while IFS=',' read -r image_name version; do
  # Trim whitespace from both fields
  image_name=$(echo "$image_name" | xargs)
  version=$(echo "$version" | xargs)

  # Add the JSON object to the array
  json_objects+=("    {\"$image_name\": \"$version\"}")
done < "$input_csv"

(IFS=$'\n'; echo "${json_objects[*]}" | sed '$!s/$/,/' >> "$output_json")

echo ']}' >> "$output_json"

echo "JSON file created: $output_json"