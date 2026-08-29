#!/usr/bin/env bash

set -euo pipefail

__wrap__() {
  VERSION=${DEVKIT_VERSION:-latest}
  PREFIX=${DEVKIT_PREFIX:-${HOME}/.local}
  REPO=tinywaves/devkit

  case "$(uname -s)" in
    Darwin)
      PLATFORM=macOS
      ;;
    Linux)
      PLATFORM=linux
      ;;
    *)
      printf '%s\n' 'error: unsupported operating system; only macOS and Linux are supported.' >&2
      exit 1
      ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64)
      ARCH=x86_64
      ;;
    arm64|aarch64)
      ARCH=arm64
      ;;
    *)
      printf '%s\n' 'error: unsupported architecture; only amd64 and arm64 are supported.' >&2
      exit 1
      ;;
  esac

  if [[ ${VERSION} != latest && ${VERSION} != v* ]]; then
    VERSION="v${VERSION}"
  fi

  ASSET="devkit_${PLATFORM}_${ARCH}.tar.gz"
  if [[ ${VERSION} == latest ]]; then
    BASE_URL="https://github.com/${REPO}/releases/latest/download"
  else
    BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
  fi

  if ! command -v curl >/dev/null 2>&1; then
    printf '%s\n' "error: curl is required but was not found on PATH." >&2
    exit 1
  fi

  if ! command -v tar >/dev/null 2>&1; then
    printf '%s\n' "error: tar is required but was not found on PATH." >&2
    exit 1
  fi

  if command -v sha256sum >/dev/null 2>&1; then
    CHECKSUM_COMMAND=sha256sum
  elif command -v shasum >/dev/null 2>&1; then
    CHECKSUM_COMMAND=shasum
  else
    printf '%s\n' 'error: sha256sum or shasum is required to verify the download.' >&2
    exit 1
  fi

  TEMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/devkit-install.XXXXXXXX")
  trap 'rm -rf "${TEMP_DIR}"' EXIT

  ARCHIVE_PATH="${TEMP_DIR}/${ASSET}"
  CHECKSUMS_PATH="${TEMP_DIR}/checksums.txt"

  printf 'Downloading devkit (%s) for %s/%s...\n' "${VERSION}" "${PLATFORM}" "${ARCH}"
  curl --fail --silent --show-error --location --retry 3 \
    "${BASE_URL}/${ASSET}" --output "${ARCHIVE_PATH}"
  curl --fail --silent --show-error --location --retry 3 \
    "${BASE_URL}/checksums.txt" --output "${CHECKSUMS_PATH}"

  EXPECTED_CHECKSUM=$(awk -v asset="${ASSET}" '$2 == asset { print $1; exit }' "${CHECKSUMS_PATH}")
  if [[ -z ${EXPECTED_CHECKSUM} ]]; then
    printf 'error: checksum for %s was not found.\n' "${ASSET}" >&2
    exit 1
  fi

  if [[ ${CHECKSUM_COMMAND} == sha256sum ]]; then
    ACTUAL_CHECKSUM=$(sha256sum "${ARCHIVE_PATH}" | awk '{print $1}')
  else
    ACTUAL_CHECKSUM=$(shasum -a 256 "${ARCHIVE_PATH}" | awk '{print $1}')
  fi

  if [[ ${ACTUAL_CHECKSUM} != ${EXPECTED_CHECKSUM} ]]; then
    printf '%s\n' 'error: checksum verification failed.' >&2
    exit 1
  fi

  tar -xzf "${ARCHIVE_PATH}" -C "${TEMP_DIR}"
  if [[ ! -f ${TEMP_DIR}/devkit ]]; then
    printf '%s\n' 'error: downloaded archive does not contain the devkit binary.' >&2
    exit 1
  fi

  DESTINATION="${PREFIX}/bin/devkit"
  mkdir -p "${PREFIX}/bin"
  install -m 0755 "${TEMP_DIR}/devkit" "${DESTINATION}"

  printf 'devkit was installed successfully to %s\n' "${DESTINATION}"
  if [[ ":${PATH}:" != *":${PREFIX}/bin:"* ]]; then
    printf 'Add %s to PATH to run devkit directly.\n' "${PREFIX}/bin"
  fi
}

__wrap__
