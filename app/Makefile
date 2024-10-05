build-dir:
	mkdir -p .build

GIT_SHA := $(shell git rev-parse HEAD)
GIT_SHA_SHORT := $(shell git rev-parse --short HEAD)
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION := $(shell git describe --tags)-$(GIT_SHA_SHORT)
LDFLAGS := -s -w \
        -X 'github.com/jlewi/foyle/app/cmd.date=$(DATE)' \
        -X 'github.com/jlewi/foyle/app/cmd.version=$(subst v,,$(VERSION))' \
        -X 'github.com/jlewi/foyle/app/cmd.commit=$(GIT_SHA)'

build: build-dir
	CGO_ENABLED=0 go build -o .build/foyle -ldflags="$(LDFLAGS)" github.com/jlewi/foyle/app

build-wasm:
	GOARCH=wasm GOOS=js go build -o web/app.wasm ./pwa

tidy-go:
	gofmt -s -w .
	goimports -w .

tidy: tidy-go

lint-go:
	# golangci-lint automatically searches up the root tree for configuration files.
	golangci-lint run

lint: lint-go

test:
	go test -v ./...
