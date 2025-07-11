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
    && wget -q -O containerd.tar.gz https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-${TARGETARCH}.tar.gz \
    && tar xzf containerd.tar.gz && mv bin/containerd bin/containerd-shim-runc-v2 bin/ctr /tools/ \
    && cd .. && rm -rf download

### nerdctl
FROM download-helper AS nerdctl
ENV NERDCTL_VERSION=2.1.3
RUN mkdir download && cd download \
    && wget -q -O nerdctl.tar.gz https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-${TARGETARCH}.tar.gz \
    && tar xzf nerdctl.tar.gz && mv nerdctl /tools/ \
    && cd .. && rm -rf download

### infra-image
# use `docker buildx imagetools inspect cgr.dev/chainguard/wolfi-base:latest` to find latest sha256 of multiarch image
FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/wolfi-base@sha256:1c6a85817d3a8787e094aae474e978d4ecdf634fd65e77ab28ffae513e35cca1

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETPLATFORM

RUN apk add --no-cache \
    iproute2 nftables losetup mount umount tini

COPY --from=runc /tools/* /usr/bin/
COPY --from=containerd /tools/* /usr/bin/
COPY --from=nerdctl /tools/* /usr/bin/

# wolfi comes with iptables-legacy as the default, but we want nft only
RUN \
    for i in iptables iptables-save iptables-restore iptables-xml; do ln -f -s /usr/sbin/xtables-nft-multi /usr/sbin/$i; done \
    && for i in iptables-xml; do ln -f -s /usr/sbin/xtables-nft-multi /usr/bin/$i; done \
    && ln -s /run /var/run \
    && mkdir -p /var/lib/containerd \
    && mkdir -p /etc/unboxed \
    && mkdir -p /var/log/unboxed \
    && mkdir -p /run/netns \
    && mkdir -p /hostfs
