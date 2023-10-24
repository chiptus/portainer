#! /bin/sh

go run -v -ldflags="-X github.com/portainer/portainer-ee/pkg/liblicense.LicenseServerBaseURL=http://localhost:8080" cmd/portainer/main.go --data=./tmp/data
