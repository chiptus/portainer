set -x

mkdir -p dist

cd api
# the go get adds 8 seconds
go get -t -d -v ./...

# the build takes 2 seconds
GOOS=$1 GOARCH=$2 CGO_ENABLED=0 go build \
	--installsuffix cgo \
    --ldflags '-s -X github.com/portainer/liblicense.LicenseServerBaseURL=https://api.portainer.io' \
	-o "../dist/portainer" \
	./cmd/portainer/