#!/usr/local/bin/bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  ./script.sh <input-file> [--in-place|-i] [--output <output-file>]

Description:
  - Scans the input text file for shepherd release links of the form:
    https://devops.oci.oraclecorp.com/shepherd/projects/oke/flocks/<flock-name>/releases/<uuid>
  - For each unique match, creates a saved query and replaces the original link
    with the shortened URL https://devops.oci.oraclecorp.com/t/<id>.

Options:
  -i, --in-place        Edit the input file in place (creates a .bak backup)
  --output <file>       Write result to the specified file (cannot combine with -i)
  -h, --help            Show this help and exit

Environment:
  OPERATOR_ACCESS_TOKEN If set, used for Authorization; otherwise the script will
                        run: ssh operator-access-token.svc.ad1.r2 "generate --mode jwt"
USAGE
}

if [[ $# -lt 1 ]]; then
  usage
  exit 1
fi

input_path="$1"; shift || true
in_place=false
output_path=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -i|--in-place)
      in_place=true
      shift
      ;;
    --output)
      if [[ $# -lt 2 ]]; then
        echo "--output requires a file path" >&2
        usage
        exit 2
      fi
      output_path="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if [[ ! -f "$input_path" ]]; then
  echo "Input file not found: $input_path" >&2
  exit 1
fi

if $in_place && [[ -n "$output_path" ]]; then
  echo "Cannot combine --in-place with --output" >&2
  exit 2
fi

# Resolve output target and prepare working copy
if $in_place; then
  backup_path="${input_path}.bak"
  cp -f "$input_path" "$backup_path"
  working_file="$(mktemp "${TMPDIR:-/tmp}/shorten.XXXXXX")"
  cp -f "$input_path" "$working_file"
  final_output="$input_path"
else
  working_file="$(mktemp "${TMPDIR:-/tmp}/shorten.XXXXXX")"
  cp -f "$input_path" "$working_file"
  if [[ -n "$output_path" ]]; then
    final_output="$output_path"
  else
    final_output="${input_path}.shortened"
  fi
fi

mapping_file="${final_output}.mapping.txt"
> "$mapping_file"

# Ensure dependencies
for dep in curl jq perl grep sort mktemp; do
  if ! command -v "$dep" >/dev/null 2>&1; then
    echo "Missing dependency: $dep" >&2
    exit 1
  fi
done

# Auth token handling
if [[ -z "${OPERATOR_ACCESS_TOKEN:-}" ]]; then
  if ! command -v ssh >/dev/null 2>&1; then
    echo "ssh not found and OPERATOR_ACCESS_TOKEN not set." >&2
    exit 1
  fi
  echo "Fetching OPERATOR_ACCESS_TOKEN via ssh..." >&2
  OPERATOR_ACCESS_TOKEN=$(ssh operator-access-token.svc.ad1.r2 "generate --mode jwt") || {
    echo "Failed to fetch OPERATOR_ACCESS_TOKEN" >&2
    exit 1
  }
  export OPERATOR_ACCESS_TOKEN
fi

# Regex pattern for shepherd release URLs (oke-ccm-csi is variable; UUID strict)
url_regex='https://devops\.oci\.oraclecorp\.com/shepherd/projects/oke/flocks/[^/]+'\
'/releases/'\
'[0-9a-fA-F]{8}(-[0-9a-fA-F]{4}){3}-[0-9a-fA-F]{12}'

# Extract unique matches
mapfile -t urls < <(grep -Eo "$url_regex" "$input_path" | sort -u || true)

total_found=${#urls[@]}
shortened=0
failed=0

if [[ $total_found -eq 0 ]]; then
  # No matches; just write/copy
  cp -f "$input_path" "$final_output"
  echo "No matching shepherd release links found. Output written to: $final_output"
  exit 0
fi

echo "Found $total_found unique shepherd release link(s). Shortening..." >&2

shorten_url() {
  local longUrl="$1"
  local payload
  payload=$(jq -nc --arg q "$longUrl" '{query:$q}')
  local response
  response=$(curl -s -H 'Content-Type: application/json' \
    -H 'Accept: */*' \
    -H "Authorization: bearer $OPERATOR_ACCESS_TOKEN" \
    --data "$payload" \
    --request POST \
    https://devops.oci.oraclecorp.com/api/ui-service/v1/saved-queries || true)

  local urlId
  urlId=$(printf '%s' "$response" | jq -r '.id // empty' 2>/dev/null || echo "")

  if [[ -n "$urlId" && "$urlId" != "null" ]]; then
    printf 'https://devops.oci.oraclecorp.com/t/%s' "$urlId"
    return 0
  else
    return 1
  fi
}

for longUrl in "${urls[@]}"; do
  # Skip if we already processed this URL (dedupe guard)
  if grep -Fq "$longUrl -> " "$mapping_file" 2>/dev/null; then
    continue
  fi

  if shortUrl=$(shorten_url "$longUrl"); then
    echo "$longUrl -> $shortUrl" >> "$mapping_file"

    # Replace literally using perl's \Q...\E quoting
    tmp_swap="$(mktemp "${TMPDIR:-/tmp}/shorten-swap.XXXXXX")"
    # Use a non-slash delimiter to avoid escaping slashes in replacement
    perl -0777 -pe "s|\\Q$longUrl\\E|$shortUrl|g" "$working_file" > "$tmp_swap"
    mv -f "$tmp_swap" "$working_file"

    shortened=$((shortened + 1))
    echo "Shortened: $longUrl -> $shortUrl" >&2
  else
    echo "WARN: Failed to shorten (no id) for $longUrl" >&2
    failed=$((failed + 1))
  fi
done

# Move working file to final output atomically
mv -f "$working_file" "$final_output"

if $in_place; then
  echo "Backup of original saved at: $backup_path" >&2
fi

echo "Summary: found=$total_found, shortened=$shortened, unchanged=$failed"
echo "Output written to: $final_output"
echo "Wrote mapping to: $mapping_file" >&2

exit 0