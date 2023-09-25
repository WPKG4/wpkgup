VERSION=0.0.1-alpha1

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
		-o wpkgup \
		.