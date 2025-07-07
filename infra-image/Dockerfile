# syntax=docker/dockerfile:1-labs

ARG WOLFI_DIGEST

### download-helper
FROM alpine AS download-helper
ARG TARGETPLATFORM
ARG TARGETARCH
RUN mkdir /tools
WORKDIR /tools

### runc
FROM download-helper AS runc
ARG RUNC_VERSION
RUN wget -q -O runc https://github.com/opencontainers/runc/releases/download/v${RUNC_VERSION}/runc.${TARGETARCH} \
    && chmod +x runc

### infra-image
FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/wolfi-base@$WOLFI_DIGEST

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETPLATFORM

RUN apk add --no-cache \
    iproute2 nftables losetup mount umount

# wolfi comes with iptables-legacy as the default, but we want nft only
RUN for i in iptables iptables-save iptables-restore iptables-xml; do ln -f -s /usr/sbin/xtables-nft-multi /usr/sbin/$i; done
RUN for i in iptables-xml; do ln -f -s /usr/sbin/xtables-nft-multi /usr/bin/$i; done

RUN ln -s /run /var/run

COPY --from=runc /tools/* /usr/sbin/
