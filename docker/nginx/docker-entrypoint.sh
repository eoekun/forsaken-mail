#!/bin/sh
set -eu

if [ -z "${BASIC_AUTH_USERNAME:-}" ] || [ -z "${BASIC_AUTH_PASSWORD:-}" ]; then
  echo "BASIC_AUTH_USERNAME and BASIC_AUTH_PASSWORD must be set" >&2
  exit 1
fi

htpasswd -cb /etc/nginx/.htpasswd "$BASIC_AUTH_USERNAME" "$BASIC_AUTH_PASSWORD"
