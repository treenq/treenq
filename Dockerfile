FROM golang:1.23.1-alpine AS builder

WORKDIR /app

# Disable CGO to ensure fully static binaries
ENV CGO_ENABLED=0 
ENV GOOS=linux

COPY go.mod go.mod
COPY go.sum go.sum
RUN --mount=type=cache,target=/go/pkg/mod/ go mod download -x

COPY . .

FROM builder AS dev

# Install Delve (debugger)
RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go install github.com/go-delve/delve/cmd/dlv@v1.23.0

RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go build -gcflags=all="-N -l" -o server ./cmd/server

# Set the default command to run the app with dlv for debugging
CMD ["dlv", "--listen=:40000", "--continue", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "server"]

FROM builder AS prod

# Create a non-root user and group for better security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# lsflags to strip debug info
RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go build -ldflags "-s -w" -o server ./cmd/server

FROM alpine:3.13

WORKDIR /app

COPY --from=prod /app/server server
COPY ./migrations /app/migrations

CMD ["/app/server"]
