#!/usr/bin/env bash
set -x

PLATFORM=$1
ARCH=$2

export GOPATH="/tmp/go"

binary="portainer"

mkdir -p dist
mkdir -p ${GOPATH}/src/github.com/portainer/portainer-ee

cp -R api ${GOPATH}/src/github.com/portainer/portainer-ee/api

cd 'api/cmd/portainer'

go get -t -d -v ./...

ldflags="-s -X 'github.com/portainer/liblicense.LicenseServerBaseURL=https://api.portainer.io' \
-X 'github.com/portainer/portainer-ee/api/build.BuildNumber=${BUILDNUMBER}' \
-X 'github.com/portainer/portainer-ee/api/build.ImageTag=${CONTAINER_IMAGE_TAG}' \
-X 'github.com/portainer/portainer-ee/api/build.NodejsVersion=${NODE_VERSION}' \
-X 'github.com/portainer/portainer-ee/api/build.YarnVersion=${YARN_VERSION}' \
-X 'github.com/portainer/portainer-ee/api/build.WebpackVersion=${WEBPACK_VERSION}' \
-X 'github.com/portainer/portainer-ee/api/build.GoVersion=${GO_VERSION}'"

if [ -n "${KAAS_AGENT_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/kubernetes/cli.DefaultAgentVersion=$KAAS_AGENT_VERSION"
fi

if [ -n "${EKSCTL_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/cloud/eks/eksctl.DefaultEksCtlVersion=$EKSCTL_VERSION"
fi

if [ -n "${AWSAUTH_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/cloud/eks/eksctl.DefaultAwsIamAuthenticatorVersion=$AWSAUTH_VERSION"
fi

echo "$ldflags"

GOOS=${PLATFORM} GOARCH=${ARCH} CGO_ENABLED=0 go build -a --installsuffix cgo --ldflags "$ldflags"

if [ "${PLATFORM}" == 'windows' ]; then
    mv "$BUILD_SOURCESDIRECTORY/api/cmd/portainer/${binary}.exe" "$BUILD_SOURCESDIRECTORY/dist/portainer.exe"
else
    mv "$BUILD_SOURCESDIRECTORY/api/cmd/portainer/$binary" "$BUILD_SOURCESDIRECTORY/dist/portainer"
fi
