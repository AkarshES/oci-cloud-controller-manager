#!/bin/bash
echo "Validating release validator POP version"
set -euo pipefail

dnf install -y jq

# Paths
manifest="MANIFEST.csv"
images_auto_tfvars_file="./shepherd/limits/flock_structure/images.auto.tfvars.json"
images_auto_tfvars_file_generated="/tmp/images.auto.tfvars.json"

# Validate required files
for f in "$manifest" "$images_auto_tfvars_file"; do
  [[ -f $f ]] || { >&2 echo "File not found: $f"; exit 1; }
done

(source ./shepherd/limits/scripts/gen_images_tfvars.sh MANIFEST.csv) > $images_auto_tfvars_file_generated

if [ ! -f "$images_auto_tfvars_file_generated" ]; then
    echo "Error: Failed to generate images auto tfvars json file."
    exit 1
fi

echo "Generated images auto tfvars json file : "
cat $images_auto_tfvars_file_generated

normalised_images_auto_tfvars_file=$(jq -S '.images |= sort_by(.name)' "$images_auto_tfvars_file")
normalised_images_auto_tfvars_generated_file=$(jq -S '.images |= sort_by(.name)' "$images_auto_tfvars_file_generated")
if [ "$normalised_images_auto_tfvars_file" = "$normalised_images_auto_tfvars_generated_file" ]; then
    echo "Verified /shepherd/limits/flock_structure/images.auto.tfvars.json"
    exit 0
else
    echo "Error: /shepherd/limits/flock_structure/images.auto.tfvars.json does not contain all the images from MANIFEST.csv"
    exit 1
fi

