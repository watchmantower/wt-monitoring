#!/usr/bin/env bash

set -euo pipefail

APP_NAME="wt-warden"
DOWNLOADS_BASE_URL="${DOWNLOADS_BASE_URL:-https://downloads.watchmantower.com/wt-warden}"
INSTALL_PATH="/usr/local/bin/${APP_NAME}"
CONFIG_DIR="/etc/watchmantower"
ENV_FILE="${CONFIG_DIR}/warden.env"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"

SERVER_ID=""
API_KEY=""
API_URL="https://metric.watchmantower.com"
VERSION=""

usage() {
  cat <<'EOF'
Usage: install.sh --server-id <id> --api-key <key> [--api-url <url>] [--version <version>]

Required:
  --server-id     Watchman Tower server id
  --api-key       Watchman Tower API key

Optional:
  --api-url       Metrics API URL (default: https://metric.watchmantower.com)
  --version       Release version, for example 1.4.0 or v1.4.0
EOF
}

log() {
  printf '[install] %s\n' "$1"
}

fail() {
  printf '[install] error: %s\n' "$1" >&2
  exit 1
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    fail "Please run this installer as root or with sudo."
  fi
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "Missing required command: $1"
}

normalize_version_tag() {
  local version="$1"
  [[ -z "$version" ]] && fail "A release version is required."
  if [[ "$version" == v* ]]; then
    printf '%s' "$version"
  else
    printf 'v%s' "$version"
  fi
}

detect_arch() {
  local uname_arch
  uname_arch="$(uname -m)"
  case "$uname_arch" in
    x86_64|amd64) printf 'amd64' ;;
    aarch64|arm64) printf 'arm64' ;;
    *) fail "Unsupported architecture: $uname_arch" ;;
  esac
}

download_binary() {
  local arch="$1"
  local version_tag="$2"
  local binary_url="${DOWNLOADS_BASE_URL}/releases/${version_tag}/${APP_NAME}_linux_${arch}"

  log "Downloading ${APP_NAME} ${version_tag} (${arch})"
  curl --fail --show-error --location "${binary_url}" --output "${INSTALL_PATH}"
  chmod 0755 "${INSTALL_PATH}"
}

write_env_file() {
  mkdir -p "${CONFIG_DIR}"
  cat > "${ENV_FILE}" <<EOF
WARDEN_SERVER_ID=${SERVER_ID}
WARDEN_API_KEY=${API_KEY}
WARDEN_API_URL=${API_URL}
EOF
  chmod 0600 "${ENV_FILE}"
}

write_service_file() {
  cat > "${SERVICE_FILE}" <<EOF
[Unit]
Description=Watchman Tower Warden Service
After=network.target

[Service]
Type=simple
EnvironmentFile=${ENV_FILE}
ExecStart=/bin/sh -c '${INSTALL_PATH} --server_id="\$WARDEN_SERVER_ID" --api_key="\$WARDEN_API_KEY" --api_url="\$WARDEN_API_URL"'
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF
}

start_service() {
  systemctl daemon-reload
  systemctl enable "${APP_NAME}" >/dev/null 2>&1 || true
  systemctl restart "${APP_NAME}"
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --server-id)
        SERVER_ID="${2:-}"
        shift 2
        ;;
      --api-key)
        API_KEY="${2:-}"
        shift 2
        ;;
      --api-url)
        API_URL="${2:-}"
        shift 2
        ;;
      --version)
        VERSION="${2:-}"
        shift 2
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        fail "Unknown argument: $1"
        ;;
    esac
  done
}

main() {
  parse_args "$@"

  [[ -n "${SERVER_ID}" ]] || fail "--server-id is required."
  [[ -n "${API_KEY}" ]] || fail "--api-key is required."
  [[ -n "${VERSION}" ]] || fail "--version is required."

  require_root
  require_command curl
  require_command systemctl

  [[ "$(uname -s)" == "Linux" ]] || fail "This installer currently supports Linux only."

  local arch version_tag
  arch="$(detect_arch)"
  version_tag="$(normalize_version_tag "${VERSION}")"

  download_binary "${arch}" "${version_tag}"
  write_env_file
  write_service_file
  start_service

  log "Warden installation complete."
  log "Service status: systemctl status ${APP_NAME}"
  log "Logs: journalctl -u ${APP_NAME} -f"
}

main "$@"
