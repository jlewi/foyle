build-dir:
	mkdir -p .build

build: build-dir
	CGO_ENABLED=0 go build -o .build/foyle github.com/jlewi/foyle/app

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
