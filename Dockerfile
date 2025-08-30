# syntax=docker/dockerfile:1.6

##############################
# Stage 1: Build Go tool
##############################
FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /src

# Cache deps
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Build the Go tool
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} \
    go build -o /out/kubeinit ./cmd/kubeinit

##############################
# Stage 2: Final image
##############################
FROM debian:bookworm-slim

ARG TARGETARCH
ARG HELM_VERSION=v3.15.2
ARG HELMFILE_VERSION=v0.165.0
ARG KUBECTL_VERSION=v1.30.3
ARG AWSCLI_VERSION=2.17.38

# Set default environment variables
ENV CLOUD_PROVIDER="aws"

# Base deps
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates \
      curl \
      unzip \
      bash \
      tini \
    && rm -rf /var/lib/apt/lists/*

################################
# Install Helm
################################
RUN case "${TARGETARCH}" in \
      amd64)  ARCH="amd64" ;; \
      arm64)  ARCH="arm64" ;; \
      arm)    ARCH="arm" ;; \
      *) echo "Unsupported arch: ${TARGETARCH}" && exit 1 ;; \
    esac \
 && curl -fsSL https://get.helm.sh/helm-${HELM_VERSION}-linux-${ARCH}.tar.gz \
    | tar -xz -C /tmp \
 && mv /tmp/linux-${ARCH}/helm /usr/local/bin/helm \
 && rm -rf /tmp/linux-${ARCH}

################################
# Install Helmfile
################################
RUN case "${TARGETARCH}" in \
      amd64)  ARCH="amd64" ;; \
      arm64)  ARCH="arm64" ;; \
      arm)    ARCH="arm" ;; \
      *) echo "Unsupported arch: ${TARGETARCH}" && exit 1 ;; \
    esac \
 && curl -fsSL https://github.com/helmfile/helmfile/releases/download/${HELMFILE_VERSION}/helmfile_${HELMFILE_VERSION#v}_linux_${ARCH}.tar.gz \
    | tar -xz -C /usr/local/bin helmfile

################################
# Install kubectl
################################
RUN case "${TARGETARCH}" in \
      amd64)  ARCH="amd64" ;; \
      arm64)  ARCH="arm64" ;; \
      arm)    ARCH="arm" ;; \
      *) echo "Unsupported arch: ${TARGETARCH}" && exit 1 ;; \
    esac \
 && curl -fsSL -o /usr/local/bin/kubectl https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/${ARCH}/kubectl \
 && chmod +x /usr/local/bin/kubectl

################################
# Install AWS CLI v2
################################
RUN case "${TARGETARCH}" in \
      amd64)  ARCH="x86_64" ;; \
      arm64)  ARCH="aarch64" ;; \
      *) echo "Unsupported arch: ${TARGETARCH}" && exit 1 ;; \
    esac \
 && curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-${ARCH}-${AWSCLI_VERSION}.zip" -o "awscliv2.zip" \
 && unzip awscliv2.zip \
 && ./aws/install --bin-dir /usr/local/bin --install-dir /usr/local/aws-cli --update \
 && rm -rf aws awscliv2.zip

################################
# Copy Go tool
################################
COPY --from=builder /out/kubeinit /usr/local/bin/kubeinit

################################
# OCI recommended labels
################################
LABEL org.opencontainers.image.title="A kubeinit Image" \
      org.opencontainers.image.description="Container with kubeinit, helmfile, helm, kubectl, and AWS CLI" \
      org.opencontainers.image.url="https://github.com/sprokhorov/kubeinit" \
      org.opencontainers.image.source="https://github.com/sprokhorov/kubeinit" \
      org.opencontainers.image.version="0.1.1" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.vendor="Sergey Prokhorov"

################################
# Security: non-root user
################################
RUN useradd -u 10001 -m appuser && mkdir /app && chown appuser /app
USER appuser
WORKDIR /app

ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["kubeinit"]
