VERSION=1.0.0

LDFLAGS = -X wpkg.dev/wpkgup/config.Version=$(VERSION)

all: build

dev:
	go run \
		-X '$(LDFLAGS)'
		.

build:
	go build \
		-v \
		-ldflags '$(LDFLAGS)' \
		.