# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22
ARG GOLANGCI_LINT_VERSION=v1.59.0
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS base
WORKDIR /src
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download -x

FROM base AS build-server
ARG TARGETOS
ARG TARGETARCH
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,target=. \
    GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /bin/app .

FROM gcr.io/distroless/base-debian12 AS app
COPY --from=build-server /bin/app /bin/
ENTRYPOINT [ "/bin/app" ]

FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION} as lint
WORKDIR /test
RUN --mount=type=bind,target=. \
    golangci-lint run
