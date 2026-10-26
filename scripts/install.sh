#!/bin/sh
set -eu

# OnboardCI POSIX installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Zerxen-dev/OnboardCI/main/scripts/install.sh | sh

REPO="Zerxen-dev/OnboardCI"
BINARY="onboardci"

echo "Installing ${BINARY}..."

# 1. Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  linux*)  OS="linux" ;;
  darwin*) OS="darwin" ;;
  *)
    echo "Error: Unsupported operating system '${OS}'. Please build from source." >&2
    exit 1
    ;;
esac

# 2. Detect Architecture
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64|amd64)   ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)
    echo "Error: Unsupported architecture '${ARCH}'. Please build from source." >&2
    exit 1
    ;;
esac

# 3. Fetch latest release version from GitHub API
RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
RELEASE_JSON=""
if command -v curl >/dev/null 2>&1; then
  RELEASE_JSON="$(curl -sSL -H "Accept: application/vnd.github.v3+json" "${RELEASE_URL}" || true)"
elif command -v wget >/dev/null 2>&1; then
  RELEASE_JSON="$(wget -qO- "${RELEASE_URL}" || true)"
else
  echo "Error: Neither curl nor wget found in PATH." >&2
  exit 1
fi

TAG_NAME="$(echo "${RELEASE_JSON}" | grep -m1 '"tag_name":' | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/')"

if [ -z "${TAG_NAME}" ]; then
  # Fallback to fetching all releases if latest is a prerelease
  RELEASES_URL="https://api.github.com/repos/${REPO}/releases"
  if command -v curl >/dev/null 2>&1; then
    RELEASE_JSON="$(curl -sSL -H "Accept: application/vnd.github.v3+json" "${RELEASES_URL}" || true)"
  else
    RELEASE_JSON="$(wget -qO- "${RELEASES_URL}" || true)"
  fi
  TAG_NAME="$(echo "${RELEASE_JSON}" | grep -m1 '"tag_name":' | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/')"
fi

if [ -z "${TAG_NAME}" ]; then
  echo "Error: Unable to determine latest release tag from GitHub." >&2
  exit 1
fi

VERSION="${TAG_NAME#v}"
ASSET_NAME="onboardci_${VERSION}_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG_NAME}/${ASSET_NAME}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${TAG_NAME}/checksums.txt"

# 4. Create temporary working directory
TMPDIR="$(mktemp -d 2>/dev/null || mktemp -d -t 'onboardci')"
cleanup() {
  rm -rf "${TMPDIR}"
}
trap cleanup EXIT INT TERM

echo "Downloading ${ASSET_NAME} (${TAG_NAME})..."
if command -v curl >/dev/null 2>&1; then
  curl -sSL -o "${TMPDIR}/${ASSET_NAME}" "${DOWNLOAD_URL}"
  curl -sSL -o "${TMPDIR}/checksums.txt" "${CHECKSUMS_URL}"
else
  wget -qO "${TMPDIR}/${ASSET_NAME}" "${DOWNLOAD_URL}"
  wget -qO "${TMPDIR}/checksums.txt" "${CHECKSUMS_URL}"
fi

# 5. Verify checksum
EXPECTED_SHA="$(grep "${ASSET_NAME}" "${TMPDIR}/checksums.txt" | awk '{print $1}')"
if [ -z "${EXPECTED_SHA}" ]; then
  echo "Error: Checksum for ${ASSET_NAME} not found in checksums.txt" >&2
  exit 1
fi

ACTUAL_SHA=""
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL_SHA="$(sha256sum "${TMPDIR}/${ASSET_NAME}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL_SHA="$(shasum -a 256 "${TMPDIR}/${ASSET_NAME}" | awk '{print $1}')"
else
  echo "Warning: sha256sum / shasum not available, skipping integrity verification."
  ACTUAL_SHA="${EXPECTED_SHA}"
fi

if [ "${EXPECTED_SHA}" != "${ACTUAL_SHA}" ]; then
  echo "Error: Checksum mismatch for ${ASSET_NAME}!" >&2
  echo "Expected: ${EXPECTED_SHA}" >&2
  echo "Got:      ${ACTUAL_SHA}" >&2
  exit 1
fi

# 6. Extract binary
tar -xzf "${TMPDIR}/${ASSET_NAME}" -C "${TMPDIR}"

# 7. Determine installation destination
INSTALL_DIR="/usr/local/bin"
USE_SUDO=0

if [ ! -w "${INSTALL_DIR}" ]; then
  if [ "$(id -u)" -eq 0 ]; then
    INSTALL_DIR="/usr/local/bin"
  elif command -v sudo >/dev/null 2>&1 && sudo -n true 2>/dev/null; then
    USE_SUDO=1
  else
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "${INSTALL_DIR}"
  fi
fi

if [ "${USE_SUDO}" -eq 1 ]; then
  sudo cp "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  sudo chmod +x "${INSTALL_DIR}/${BINARY}"
else
  cp "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"
fi

echo "Successfully installed ${BINARY} ${TAG_NAME} to ${INSTALL_DIR}/${BINARY}!"

case ":${PATH}:" in
  *":${INSTALL_DIR}:"*) ;;
  *)
    echo ""
    echo "Note: '${INSTALL_DIR}' is not in your PATH."
    echo "Add it by adding this line to your ~/.bashrc or ~/.zshrc:"
    echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
    ;;
esac
