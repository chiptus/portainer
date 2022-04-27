#!/usr/bin/env bash

PLATFORM=$1
ARCH=$2

export GOPATH="/tmp/go"

binary="portainer"

mkdir -p dist
mkdir -p ${GOPATH}/src/github.com/portainer/portainer-ee

cp -R api ${GOPATH}/src/github.com/portainer/portainer-ee/api

cd 'api/cmd/portainer'

go get -t -d -v ./...

ldflags='-s -X github.com/portainer/liblicense.LicenseServerBaseURL=https://api.portainer.io'
if [ -n "${KAAS_AGENT_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/kubernetes/cli.DefaultAgentVersion=$KAAS_AGENT_VERSION"
fi
echo "$ldflags"

GOOS=${PLATFORM} GOARCH=${ARCH} CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags "$ldflags"

if [ "${PLATFORM}" == 'windows' ]; then
    mv "$BUILD_SOURCESDIRECTORY/api/cmd/portainer/${binary}.exe" "$BUILD_SOURCESDIRECTORY/dist/portainer.exe"
else
    mv "$BUILD_SOURCESDIRECTORY/api/cmd/portainer/$binary" "$BUILD_SOURCESDIRECTORY/dist/portainer"
fi
