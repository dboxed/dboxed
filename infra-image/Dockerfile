# syntax=docker/dockerfile:1-labs

### download-helper
FROM alpine AS download-helper
ARG TARGETPLATFORM
ARG TARGETARCH
RUN mkdir /tools
WORKDIR /tools

### runc
FROM download-helper AS runc
ENV RUNC_VERSION=1.3.0
RUN wget -q -O runc https://github.com/opencontainers/runc/releases/download/v${RUNC_VERSION}/runc.${TARGETARCH} \
    && chmod +x runc

### containerd
FROM download-helper AS containerd
ENV CONTAINERD_VERSION=2.1.3
RUN mkdir download && cd download \
    && wget -q -O download.tar.gz https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-static-${CONTAINERD_VERSION}-linux-${TARGETARCH}.tar.gz \
    && tar xzf download.tar.gz && mv bin/containerd bin/containerd-shim-runc-v2 bin/ctr /tools/ \
    && cd .. && rm -rf download

### nerdctl
FROM download-helper AS nerdctl
ENV NERDCTL_VERSION=2.1.3
RUN mkdir download && cd download \
    && wget -q -O download.tar.gz https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-${TARGETARCH}.tar.gz \
    && tar xzf download.tar.gz && mv nerdctl /tools/ \
    && cd .. && rm -rf download

### cni
FROM download-helper AS cni
ENV CNI_VERSION=1.7.1
RUN mkdir download && cd download \
    && wget -q -O download.tar.gz https://github.com/containernetworking/plugins/releases/download/v${CNI_VERSION}/cni-plugins-linux-${TARGETARCH}-v${CNI_VERSION}.tgz \
    && tar xzf download.tar.gz && mv bridge firewall portmap tuning host-local /tools/ \
    && cd .. && rm -rf download

### infra-image
FROM --platform=$TARGETPLATFORM alpine:3.22.0

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETPLATFORM

RUN apk add --no-cache \
    iproute2 iptables tini ca-certificates

COPY --from=runc /tools/* /usr/bin/
COPY --from=containerd /tools/* /usr/bin/
COPY --from=nerdctl /tools/* /usr/bin/
COPY --from=cni /tools/* /opt/cni/bin/

RUN mkdir -p /var/lib/containerd \
    && mkdir -p /etc/unboxed \
    && mkdir -p /var/log/unboxed \
    && mkdir -p /run/netns \
    && mkdir -p /hostfs
