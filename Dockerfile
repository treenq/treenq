# Build binary stage
FROM golang:1.22.3-alpine AS builder

WORKDIR /app

# Add a debugger, the flag must be used only in e2e tests
ARG DEBUG=false
RUN if [ $DEBUG = true ]; then go install github.com/go-delve/delve/cmd/dlv@latest; fi

# adding modules files before building in order to avoid cache invalidation
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .

# Build binary
# remove optimization for better debugging experience if DEBUG is true
RUN CGO_ENABLED=0 GOOS=linux go build -gcflags="$(if [ \"$DEBUG\" = \"true\" ]; then echo 'all=-N -l'; else echo ''; fi)" -o server ./cmd/server

# Run binary stage
FROM alpine:3.13

WORKDIR /app

COPY --from=builder /app/server server
COPY --from=builder /go/bin/dlv* /
COPY migrations migrations

RUN chmod +x /app/server

CMD ["/app/server"]
