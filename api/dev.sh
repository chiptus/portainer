#! /bin/sh

go run -v -ldflags="-X github.com/portainer/liblicense/v3.LicenseServerBaseURL=http://localhost:8080" cmd/portainer/main.go --data=./tmp/data
