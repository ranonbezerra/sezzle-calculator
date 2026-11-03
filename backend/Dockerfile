# syntax=docker/dockerfile:1

# Must match the go directive in go.mod. GOTOOLCHAIN=local makes the build
# fail instead of silently downloading another toolchain if they drift.
ARG GO_VERSION=1.27.3

FROM golang:${GO_VERSION}-alpine AS build
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static:nonroot
COPY --from=build /out/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
