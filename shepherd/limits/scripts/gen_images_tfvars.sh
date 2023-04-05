#!/bin/bash
set -e

manifest="../../MANIFEST.csv"

echo "{\"images\": ["

awk ' \
  BEGIN { SEP=""; FS=",";} \
  /.\/*/{ \
    gsub(/^[ \t]+|[ \t]+$/, "", $1); \
    gsub(/^[ \t]+|[ \t]+$/, "", $2); \
    gsub("\\.", "_DOT_", $2); \
    printf "%s  {\"name\": \"%s__%s\", \"location\": \"%s\"}", SEP, $1, $2, $1; \
    SEP=",\n"; \
  } \
  END {printf "\n"}' ${manifest}

echo "]}"
