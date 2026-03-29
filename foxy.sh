#!/usr/bin/env bash

set -euo pipefail

if [ "$#" -ne 1 ]; then
  printf 'Usage: %s <payload-file>\n' "$0" >&2
  exit 1
fi

payload_file="$1"

if [ ! -f "$payload_file" ]; then
  printf 'Error: file not found: %s\n' "$payload_file" >&2
  exit 1
fi

curl --no-buffer -X POST "http://localhost:6942/v1" \
  -H "Accept: application/json" \
  -H "Content-Type: application/json" \
  -d @"$payload_file"
