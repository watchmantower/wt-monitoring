#!/usr/bin/env bash

set -euo pipefail

APP_NAME="wt-warden"
LEGACY_APP_NAME="wt-monitoring"
DOWNLOADS_BASE_URL="${DOWNLOADS_BASE_URL:-https://downloads.watchmantower.com/wt-warden}"
INSTALL_PATH="/usr/local/bin/${APP_NAME}"
CONFIG_DIR="/etc/watchmantower"
ENV_FILE="${CONFIG_DIR}/warden.env"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"
LEGACY_SERVICE_FILE="/etc/systemd/system/${LEGACY_APP_NAME}.service"

VERSION=""

usage() {
  cat <<'EOF'
Usage: upgrade.sh --version <version>

Optional:
  --version       Release version, for example 1.4.0 or v1.4.0
EOF
}

log() {
  printf '[upgrade] %s\n' "$1"
}

fail() {
  printf '[upgrade] error: %s\n' "$1" >&2
  exit 1
}

require_root() {
  if [[ "${EUID}" -ne 0 ]]; then
    fail "Please run this upgrader as root or with sudo."
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

ensure_env_file() {
  if [[ -f "${ENV_FILE}" ]]; then
    return
  fi

  mkdir -p "${CONFIG_DIR}"

  if [[ -f "${LEGACY_SERVICE_FILE}" ]]; then
    local exec_line server_id api_key api_url
    exec_line="$(grep -E '^ExecStart=' "${LEGACY_SERVICE_FILE}" || true)"
    server_id="$(printf '%s' "${exec_line}" | sed -n 's/.*--server_id=\([^ ]*\).*/\1/p')"
    api_key="$(printf '%s' "${exec_line}" | sed -n 's/.*--api_key=\([^ ]*\).*/\1/p')"
    api_url="$(printf '%s' "${exec_line}" | sed -n 's/.*--api_url="\([^"]*\)".*/\1/p')"

    [[ -n "${server_id}" ]] || fail "Could not migrate server_id from legacy service."
    [[ -n "${api_key}" ]] || fail "Could not migrate api_key from legacy service."
    [[ -n "${api_url}" ]] || api_url="https://metric.watchmantower.com"

    cat > "${ENV_FILE}" <<EOF
WARDEN_SERVER_ID=${server_id}
WARDEN_API_KEY=${api_key}
WARDEN_API_URL=${api_url}
EOF
    chmod 0600 "${ENV_FILE}"
    return
  fi

  fail "No ${ENV_FILE} found and no legacy service available to migrate from."
}

download_binary() {
  local arch="$1"
  local version_tag="$2"
  local binary_url="${DOWNLOADS_BASE_URL}/releases/${version_tag}/${APP_NAME}_linux_${arch}"

  log "Downloading ${APP_NAME} ${version_tag} (${arch})"
  curl --fail --show-error --location "${binary_url}" --output "${INSTALL_PATH}"
  chmod 0755 "${INSTALL_PATH}"
}

write_service_file() {
  cat > "${SERVICE_FILE}" <<EOF
[Unit]
Description=Watchman Tower Warden Service
After=network.target

[Service]
Type=simple
EnvironmentFile=${ENV_FILE}
ExecStart=/bin/sh -c '${INSTALL_PATH} --server_id="$$WARDEN_SERVER_ID" --api_key="$$WARDEN_API_KEY" --api_url="$$WARDEN_API_URL"'
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
EOF
}

stop_existing_services() {
  systemctl stop "${LEGACY_APP_NAME}" >/dev/null 2>&1 || true
  systemctl stop "${APP_NAME}" >/dev/null 2>&1 || true
}

start_service() {
  systemctl daemon-reload
  systemctl disable "${LEGACY_APP_NAME}" >/dev/null 2>&1 || true
  systemctl enable "${APP_NAME}" >/dev/null 2>&1 || true
  systemctl restart "${APP_NAME}"
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
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

  [[ -n "${VERSION}" ]] || fail "--version is required."

  require_root
  require_command curl
  require_command systemctl

  [[ "$(uname -s)" == "Linux" ]] || fail "This upgrader currently supports Linux only."

  local arch version_tag
  arch="$(detect_arch)"
  version_tag="$(normalize_version_tag "${VERSION}")"

  stop_existing_services
  ensure_env_file
  download_binary "${arch}" "${version_tag}"
  write_service_file
  start_service

  log "Warden upgrade complete."
  log "Service status: systemctl status ${APP_NAME}"
  log "Logs: journalctl -u ${APP_NAME} -f"
}

main "$@"
