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

# 3. Determine release tag (allow VERSION env override, otherwise fetch from GitHub API)
TAG_NAME="${VERSION:-}"
if [ -z "${TAG_NAME}" ]; then
  RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
  if command -v curl >/dev/null 2>&1; then
    RELEASE_JSON="$(curl -sSL -H "Accept: application/vnd.github.v3+json" "${RELEASE_URL}" 2>/dev/null || true)"
  elif command -v wget >/dev/null 2>&1; then
    RELEASE_JSON="$(wget -qO- "${RELEASE_URL}" 2>/dev/null || true)"
  fi
  TAG_NAME="$(echo "${RELEASE_JSON:-}" | grep -m1 '"tag_name":' | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/' || true)"
fi

if [ -z "${TAG_NAME}" ]; then
  RELEASES_URL="https://api.github.com/repos/${REPO}/releases"
  if command -v curl >/dev/null 2>&1; then
    RELEASE_JSON="$(curl -sSL -H "Accept: application/vnd.github.v3+json" "${RELEASES_URL}" 2>/dev/null || true)"
  elif command -v wget >/dev/null 2>&1; then
    RELEASE_JSON="$(wget -qO- "${RELEASES_URL}" 2>/dev/null || true)"
  fi
  TAG_NAME="$(echo "${RELEASE_JSON:-}" | grep -m1 '"tag_name":' | sed -E 's/.*"tag_name":[[:space:]]*"([^"]+)".*/\1/' || true)"
fi

if [ -z "${TAG_NAME}" ]; then
  echo "Error: Unable to determine release tag from GitHub." >&2
  exit 1
fi

# 4. Create temporary directory
TMPDIR="$(mktemp -d 2>/dev/null || mktemp -d -t 'onboardci')"
cleanup() {
  rm -rf "${TMPDIR}"
}
trap cleanup EXIT INT TERM

# 5. Fetch checksums.txt
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${TAG_NAME}/checksums.txt"
if command -v curl >/dev/null 2>&1; then
  curl -sSL -o "${TMPDIR}/checksums.txt" "${CHECKSUMS_URL}"
else
  wget -qO "${TMPDIR}/checksums.txt" "${CHECKSUMS_URL}"
fi

# 6. Locate asset for OS/ARCH from checksums.txt
MATCH_LINE="$(grep -E "[[:space:]]+.*${OS}.*${ARCH}\.tar\.gz$" "${TMPDIR}/checksums.txt" | head -n 1)"
if [ -z "${MATCH_LINE}" ]; then
  echo "Error: No matching archive found for ${OS}/${ARCH} in ${TAG_NAME} checksums." >&2
  exit 1
fi

EXPECTED_SHA="$(echo "${MATCH_LINE}" | awk '{print $1}')"
ASSET_NAME="$(echo "${MATCH_LINE}" | awk '{print $2}')"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${TAG_NAME}/${ASSET_NAME}"

echo "Downloading ${ASSET_NAME} (${TAG_NAME})..."
if command -v curl >/dev/null 2>&1; then
  curl -sSL -o "${TMPDIR}/${ASSET_NAME}" "${DOWNLOAD_URL}"
else
  wget -qO "${TMPDIR}/${ASSET_NAME}" "${DOWNLOAD_URL}"
fi

# 7. Verify checksum integrity
ACTUAL_SHA=""
if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL_SHA="$(sha256sum "${TMPDIR}/${ASSET_NAME}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL_SHA="$(shasum -a 256 "${TMPDIR}/${ASSET_NAME}" | awk '{print $1}')"
else
  echo "Error: Neither sha256sum nor shasum is available to verify archive integrity." >&2
  exit 1
fi

if [ "${EXPECTED_SHA}" != "${ACTUAL_SHA}" ]; then
  echo "Error: Checksum mismatch for ${ASSET_NAME}!" >&2
  echo "Expected: ${EXPECTED_SHA}" >&2
  echo "Got:      ${ACTUAL_SHA}" >&2
  exit 1
fi

# 8. Extract binary
tar -xzf "${TMPDIR}/${ASSET_NAME}" -C "${TMPDIR}"

# 9. Install binary
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
