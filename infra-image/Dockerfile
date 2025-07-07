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

### infra-image
# use `docker buildx imagetools inspect cgr.dev/chainguard/wolfi-base:latest` to find latest sha256 of multiarch image
FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/wolfi-base@sha256:1c6a85817d3a8787e094aae474e978d4ecdf634fd65e77ab28ffae513e35cca1

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETPLATFORM

RUN apk add --no-cache \
    iproute2 nftables losetup mount umount

# wolfi comes with iptables-legacy as the default, but we want nft only
RUN for i in iptables iptables-save iptables-restore iptables-xml; do ln -f -s /usr/sbin/xtables-nft-multi /usr/sbin/$i; done
RUN for i in iptables-xml; do ln -f -s /usr/sbin/xtables-nft-multi /usr/bin/$i; done

RUN ln -s /run /var/run

COPY --from=runc /tools/* /usr/sbin/
