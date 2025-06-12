FROM candyboobers/treenq-base AS builder

WORKDIR /app

COPY go.mod go.mod
COPY go.sum go.sum
RUN --mount=type=cache,target=/go/pkg/mod/ go mod download -x

# DEV target, has dlv 
FROM builder AS dev

# Install Delve (debugger)
RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go install github.com/go-delve/delve/cmd/dlv@v1.24.1

# Copy after installed delve in order to cache delve in the upper layer
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go build -race -gcflags=all="-N -l" -o server ./cmd/server

# Set the default command to run the app with dlv for debugging
CMD ["dlv", "--listen=:40000", "--continue", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "server"]

FROM builder AS prod

# Copy after installed delve in order to cache delve in the upper layer
COPY . .

# lsflags to strip debug info
RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go build -ldflags "-s -w" -o server ./cmd/server

CMD ["/app/server"]

# FROM alpine:3.13
#
# # Create a non-root user and group for better security
# # RUN addgroup -S appgroup && adduser -S 1001 -G appgroup
# RUN addgroup -g 1001 appgroup && adduser -D -G appgroup -u 1001 appuser
#
# WORKDIR /app
#
# USER 1001
#
# COPY --from=prod /app/server server
# COPY ./migrations /app/migrations
#
# CMD ["/app/server"]
