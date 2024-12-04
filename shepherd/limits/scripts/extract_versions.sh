#!/bin/bash

usage() {
  echo "Usage: $0 <path-to-wheelbarrow-manifest>"
  exit 1
}

if [ "$#" -ne 1 ]; then
  usage
fi

input_csv="$1"
output_json="release-validator-pop/image_versions.json"

if [ ! -f "$input_csv" ] || [ ! -r "$input_csv" ]; then
  echo "Error: Input file does not exist or cannot be read."
  exit 1
fi

awk -F, '{gsub(/^ *| *$/,"",$1); gsub(/^ *| *$/,"",$2); print "{\""$1"\": \""$2"\"}"}' "$input_csv" | jq -s '. | {images: .}' > "$output_json"

echo "JSON file created: $output_json"