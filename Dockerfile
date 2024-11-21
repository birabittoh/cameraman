# syntax=docker/dockerfile:1

FROM golang:1.22.3-alpine AS builder

WORKDIR /build

# Install gcc and musl-dev
RUN apk add --no-cache gcc musl-dev

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Transfer source code
COPY templates ./templates
COPY *.go ./

# Build
RUN CGO_ENABLED=1 go build -ldflags='-s -w' -trimpath -o /dist/app
RUN ldd /dist/app | tr -s [:blank:] '\n' | grep ^/ | xargs -I % install -D % /dist/%
RUN ln -s ld-musl-x86_64.so.1 /dist/lib/libc.musl-x86_64.so.1

# Test
FROM builder AS run-test-stage
RUN go test -v ./...

FROM scratch AS build-release-stage

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dist /

ENTRYPOINT ["/app"]
