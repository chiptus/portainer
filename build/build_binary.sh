#!/bin/sh

set -x

mkdir -p dist

cd api
# the go get adds 8 seconds
go get -t -d -v ./...


ldflags='-s -X github.com/portainer/liblicense.LicenseServerBaseURL=https://api.portainer.io'
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

# the build takes 2 seconds
GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build \
	--installsuffix cgo \
    --ldflags "$ldflags" \
	-o "../dist/portainer" \
	./cmd/portainer/
