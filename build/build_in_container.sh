#!/usr/bin/env sh

binary="portainer-$1-$2"

mkdir -p dist

ldflags='-s -X github.com/portainer/liblicense.LicenseServerBaseURL=https://api.portainer.io'
if [ -n "${KAAS_AGENT_VERSION+1}" ]; then
  ldflags=$ldflags" -X github.com/portainer/portainer-ee/api/kubernetes/cli.DefaultAgentVersion=$KAAS_AGENT_VERSION"
fi

echo "$ldflags"

docker run --rm -tv "$(pwd)/api:/src" -e BUILD_GOOS="$1" -e BUILD_GOARCH="$2" -e LDFLAGS="$ldflags" portainer/golang-builder:cross-platform /src/cmd/portainer

mv "api/cmd/portainer/$binary" dist/
#sha256sum "dist/$binary" > portainer-checksum.txt
