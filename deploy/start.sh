#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")"

export ADDR="${ADDR:-0.0.0.0:8080}"
export TRUSTED_PROXIES="${TRUSTED_PROXIES:-cloudflare}"
export POW_DIFFICULTY="${POW_DIFFICULTY:-12}"
exec ./e2ee-chat
