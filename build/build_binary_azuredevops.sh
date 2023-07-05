#!/usr/bin/env bash
set -x

PLATFORM=$1
ARCH=$2

export GOPATH="/tmp/go"
BUILD_SOURCESDIRECTORY=${BUILD_SOURCESDIRECTORY:-$(pwd)}

mkdir -p dist
mkdir -p ${GOPATH}/src/github.com/portainer/portainer-ee

cp -R api ${GOPATH}/src/github.com/portainer/portainer-ee/api

cp -r "./mustache-templates" "./dist"

cd 'api/cmd/portainer' || exit 1

go get -t -d -v ./...

ldflags="-s -X 'github.com/portainer/liblicense/v3.LicenseServerBaseURL=https://api.portainer.io' \
-X 'github.com/portainer/portainer-ee/api/build.BuildNumber=${BUILDNUMBER}' \
-X 'github.com/portainer/portainer-ee/api/build.ImageTag=${CONTAINER_IMAGE_TAG}' \
-X 'github.com/portainer/portainer-ee/api/build.NodejsVersion=${NODE_VERSION}' \
-X 'github.com/portainer/portainer-ee/api/build.YarnVersion=${YARN_VERSION}' \
-X 'github.com/portainer/portainer-ee/api/build.WebpackVersion=${WEBPACK_VERSION}' \
-X 'github.com/portainer/portainer-ee/api/build.GoVersion=${GO_VERSION}'"

if [ -n "${KAAS_AGENT_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/kubernetes/cli.DefaultAgentVersion=$KAAS_AGENT_VERSION"
fi

BINARY_VERSION_FILE="$BUILD_SOURCESDIRECTORY/binary-version.json"

EKSCTL_VERSION=$(jq -r '.eksctl' < "${BINARY_VERSION_FILE}")
if [ -n "${EKSCTL_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/cloud/eks/eksctl.DefaultEksCtlVersion=$EKSCTL_VERSION"
fi

AWSAUTH_VERSION=$(jq -r '.awsAuth' < "${BINARY_VERSION_FILE}")
if [ -n "${AWSAUTH_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/cloud/eks/eksctl.DefaultAwsIamAuthenticatorVersion=$AWSAUTH_VERSION"
fi

GOOS=${PLATFORM} GOARCH=${ARCH} CGO_ENABLED=0 go build -a -trimpath --installsuffix cgo --gcflags="-trimpath $(pwd)" --ldflags "$ldflags"

if [ "${PLATFORM}" == 'windows' ]; then
    mv "$BUILD_SOURCESDIRECTORY/api/cmd/portainer/portainer.exe" "$BUILD_SOURCESDIRECTORY/dist/portainer.exe"
else
    mv "$BUILD_SOURCESDIRECTORY/api/cmd/portainer/portainer" "$BUILD_SOURCESDIRECTORY/dist/portainer"
fi
