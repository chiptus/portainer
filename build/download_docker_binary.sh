#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 3 ]]; then
    echo "Illegal number of parameters" >&2
    exit 1
fi

PLATFORM=$1
ARCH=$2
DOCKER_VERSION=${3:1}
DOWNLOAD_FOLDER=".tmp/download"

if [[ ${PLATFORM} == "darwin" ]]; then
    PLATFORM="mac"
fi

if [[ ${ARCH} == "amd64" ]]; then
    ARCH="x86_64"
elif [[ ${ARCH} == "arm" ]]; then
    ARCH="armhf"
elif [[ ${ARCH} == "arm64" ]]; then
    ARCH="aarch64"
elif [[ ${ARCH} == "ppc64le" ]]; then
    DOCKER_VERSION="18.06.3-ce"
elif [[ ${ARCH} == "s390x" ]]; then
    DOCKER_VERSION="18.06.3-ce"
fi

rm -rf "${DOWNLOAD_FOLDER}"
mkdir -pv "${DOWNLOAD_FOLDER}"

if [[ ${PLATFORM} == "windows" ]]; then
    # if [[ ! -f "dist/docker.exe" ]]; then
    wget --tries=3 --waitretry=30 --quiet -O "${DOWNLOAD_FOLDER}/docker-binaries.zip" "https://download.docker.com/win/static/stable/${ARCH}/docker-${DOCKER_VERSION}.zip"
    unzip "${DOWNLOAD_FOLDER}/docker-binaries.zip" -d "${DOWNLOAD_FOLDER}"
    mv "${DOWNLOAD_FOLDER}/docker/docker.exe" dist/
    # fi
else
    # if [[ ! -f "dist/docker" ]]; then
    wget --tries=3 --waitretry=30 --quiet -O "${DOWNLOAD_FOLDER}/docker-binaries.tgz" "https://download.docker.com/${PLATFORM}/static/stable/${ARCH}/docker-${DOCKER_VERSION}.tgz"
    tar -xf "${DOWNLOAD_FOLDER}/docker-binaries.tgz" -C "${DOWNLOAD_FOLDER}"
    mv "${DOWNLOAD_FOLDER}/docker/docker" dist/
    # fi
fi
