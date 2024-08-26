ARG BUILD_IMAGE=golang:1.23
ARG RUNTIME_IMAGE=cgr.dev/chainguard/static:latest
FROM ${BUILD_IMAGE} as builder

# Build Args need to be after the FROM stage otherwise they don't get passed through
ARG VERSION=unknown
ARG DATE=unknown
ARG COMMIT=unknown

WORKDIR /workspace/

COPY . /workspace

WORKDIR /workspace/app

## Build
# The LDFLAG can't be specified multiple times so we use an environment variable to build it up over multiple lines
RUN LDFLAGS="-s -w -X github.com/jlewi/foyle/app/cmd/commands.version=${VERSION}" && \
    LDFLAGS="${LDFLAGS} -X github.com/jlewi/foyle/app/cmd/commands.commit=${COMMIT}" && \
    LDFLAGS="${LDFLAGS} -X github.com/jlewi/foyle/app/cmd/commands.date=${DATE}" && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on \
    go build \
    -ldflags "${LDFLAGS}" \
    -a -o foyle github.com/jlewi/foyle/app/cmd

FROM ${RUNTIME_IMAGE}

COPY --from=builder /workspace/app/foyle /
RUN chmod +x /foyle
ENTRYPOINT ["/foyle", "serve"]
