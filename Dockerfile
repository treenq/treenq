FROM ubuntu:25.04 AS builder

WORKDIR /app

RUN apt-get update
RUN apt-get -y install buildah bats btrfs-progs git go-md2man golang libapparmor-dev libglib2.0-dev libgpgme11-dev libseccomp-dev libselinux1-dev make runc skopeo libbtrfs-dev wget fuse-overlayfs && rm -rf /var/lib/apt/lists/*
RUN mkdir -p /etc/containers && \
    mkdir -p /var/lib/shared/overlay-images /var/lib/shared/overlay-layers && \
    touch /var/lib/shared/overlay-images/images.lock && \
    touch /var/lib/shared/overlay-layers/layers.lock
RUN wget -P /tmp https://go.dev/dl/go1.24.1.linux-arm64.tar.gz
RUN tar -C /usr/local -xzf "/tmp/go1.24.1.linux-arm64.tar.gz"
RUN rm "/tmp/go1.24.1.linux-arm64.tar.gz"

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

RUN mkdir -p /etc/containers/ && touch /etc/containers/registries.conf && echo 'unqualified-search-registries=["docker.io"]' > /etc/containers/registries.conf
COPY policy.json /etc/containers/policy.json
COPY storage.conf /etc/containers/storage.conf 

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

# lsflags to strip debug info
RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go build -ldflags "-s -w" -o server ./cmd/server

FROM alpine:3.13

# Create a non-root user and group for better security
# RUN addgroup -S appgroup && adduser -S 1001 -G appgroup
RUN addgroup -g 1001 appgroup && adduser -D -G appgroup -u 1001 appuser

WORKDIR /app

USER 1001

COPY --from=prod /app/server server
COPY ./migrations /app/migrations

CMD ["/app/server"]
