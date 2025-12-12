APP=serverpatcher
PKG=./cmd/serverpatcher

.PHONY: build test vet install

build:
	go build -trimpath -ldflags "-s -w" -o bin/$(APP) $(PKG)

test:
	go test ./...

vet:
	go vet ./...

install:
	sudo ./scripts/install.sh
