FROM golang:1.24.2 AS builder

WORKDIR /app

RUN apt-get update && apt-get -y install --no-install-recommends \
    buildah \
    gnupg \
    btrfs-progs \
    bats \
    libapparmor-dev \
    libglib2.0-dev \
    libgpgme11-dev \
    libseccomp-dev \
    libselinux1-dev \
    runc \
    skopeo \
    libbtrfs-dev \
    fuse-overlayfs \
    && curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get -y install --no-install-recommends nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /usr/libexec/podman /etc/containers/

RUN curl -L -o /usr/libexec/podman/netavark.gz https://github.com/containers/netavark/releases/download/v1.14.1/netavark.gz && \
    curl -L -o /usr/libexec/podman/aardvark-dns.gz https://github.com/containers/aardvark-dns/releases/download/v1.14.0/aardvark-dns.gz && \
    cd /usr/libexec/podman && \
    gunzip netavark.gz && \
    gunzip aardvark-dns.gz && \
    chmod +x netavark && \
    chmod +x aardvark-dns

RUN mkdir -p /etc/containers/ && touch /etc/containers/registries.conf && echo 'unqualified-search-registries=["docker.io"]' > /etc/containers/registries.conf

COPY policy.json storage.conf registries.conf /etc/containers/

# # Disable CGO to ensure fully static binaries
# ENV CGO_ENABLED=0
# ENV GOOS=linux

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod/ go mod download -x

FROM builder AS debugger

# Install Delve (debugger)
RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go install github.com/go-delve/delve/cmd/dlv@v1.24.1

# Copy after installed delve in order to cache delve in the upper layer
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go build -gcflags=all="-N -l" -o server ./cmd/server

FROM debugger AS dev
CMD ["dlv", "--listen=:40000", "--continue", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "server"]

FROM debugger AS e2e
CMD ["dlv", "--listen=:41000", "--continue", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "server"]

FROM builder AS prod

# Copy after installed delve in order to cache delve in the upper layer
COPY . .

# lsflags to strip debug info
RUN --mount=type=cache,target=/go/pkg/mod/ --mount=type=cache,target="/root/.cache/go-build" go build -ldflags "-s -w" -o server ./cmd/server

FROM scratch

# Create a non-root user and group for better security
# RUN addgroup -S appgroup && adduser -S 1001 -G appgroup
RUN addgroup -g 1001 appgroup && adduser -D -G appgroup -u 1001 appuser

WORKDIR /app

RUN mkdir -p /usr/libexec/podman

COPY --from=builder /usr/libexec/podman/netavark /usr/libexec/podman/
COPY --from=builder /usr/libexec/podman/aardvark-dns /usr/libexec/podman/

RUN chmod +x /usr/libexec/podman/netavark && \
    chmod +x /usr/libexec/podman/aardvark-dns

USER 1001

COPY --from=prod /app/server server
COPY ./migrations /app/migrations

CMD ["/app/server"]
