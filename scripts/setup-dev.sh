#!/usr/bin/env bash
# Optional dev shortcut — paste credentials from an existing Slack app.
# Normal flow: theirtime onboard (opens manifest URL for you).
set -euo pipefail

CONFIG_DIR="${HOME}/.config/theirtime"
ENV_FILE="${CONFIG_DIR}/.env"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

mkdir -p "${CONFIG_DIR}"

if [[ -f "${ENV_FILE}" ]]; then
  echo "Already exists: ${ENV_FILE}"
else
  cp "${ROOT}/.env.example" "${ENV_FILE}"
  echo "Created ${ENV_FILE}"
fi

echo ""
echo "Optional dev shortcut: edit ${ENV_FILE} with Client ID + Secret from an existing app."
echo "Otherwise just run: make build && ./bin/theirtime onboard"
