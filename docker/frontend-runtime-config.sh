#!/bin/sh
# Regenerates the SPA runtime config from environment at container start.
# GOSIGN_API_URL: base URL of the API, e.g. https://api.example.com/v1
# (empty = same-origin /v1).
set -e

cat > /usr/share/nginx/html/config.js <<EOF
window.__GOSIGN_API_URL__ = "${GOSIGN_API_URL:-}";
EOF
