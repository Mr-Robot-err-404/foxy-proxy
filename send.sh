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

curl --no-buffer -X POST "https://api.anthropic.com/v1/messages?beta=true" \
  -H "Accept: application/json" \
  -H "Authorization: Bearer $CLAUDE_BEARER" \
  -H "Content-Type: application/json" \
  -H "User-Agent: claude-cli/2.1.81 (external, cli)" \
  -H "X-Stainless-Arch: arm64" \
  -H "X-Stainless-Lang: js" \
  -H "X-Stainless-OS: MacOS" \
  -H "X-Stainless-Package-Version: 0.74.0" \
  -H "X-Stainless-Retry-Count: 0" \
  -H "X-Stainless-Runtime: node" \
  -H "X-Stainless-Runtime-Version: v24.3.0" \
  -H "X-Stainless-Timeout: 600" \
  -H "anthropic-beta: claude-code-20250219,oauth-2025-04-20,interleaved-thinking-2025-05-14,redact-thinking-2026-02-12,context-management-2025-06-27,prompt-caching-scope-2026-01-05,advanced-tool-use-2025-11-20,effort-2025-11-24" \
  -H "anthropic-dangerous-direct-browser-access: true" \
  -H "anthropic-version: 2023-06-01" \
  -H "x-app: cli" \
  -d @"$payload_file"
