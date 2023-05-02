#!/usr/bin/env bash
set -euo pipefail

mkdir -p dist

# populate tool versions
BUILDNUMBER="N/A"
CONTAINER_IMAGE_TAG="N/A"
NODE_VERSION="0"
YARN_VERSION="0"
WEBPACK_VERSION="0"
GO_VERSION="0"

# copy templates
cp -r "./mustache-templates" "./dist"


cd api || exit 1
# the go get adds 8 seconds
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

BINARY_VERSION_FILE="./binary-version.json"

EKSCTL_VERSION=$(jq -r '.eksctl' < "${BINARY_VERSION_FILE}")
if [ -n "${EKSCTL_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/cloud/eks/eksctl.DefaultEksCtlVersion=$EKSCTL_VERSION"
fi

AWSAUTH_VERSION=$(jq -r '.awsAuth' < "${BINARY_VERSION_FILE}")
if [ -n "${AWSAUTH_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/cloud/eks/eksctl.DefaultAwsIamAuthenticatorVersion=$AWSAUTH_VERSION"
fi

echo "$ldflags"

# the build takes 2 seconds
GOOS=${1:-$(go env GOOS)} GOARCH=${2:-$(go env GOARCH)} CGO_ENABLED=0 go build \
	-trimpath \
	--installsuffix cgo \
	--ldflags "$ldflags" \
	-o "../dist/portainer" \
	./cmd/portainer/
