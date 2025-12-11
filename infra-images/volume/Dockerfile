# syntax=docker/dockerfile:1-labs

# rootfs
FROM alpine:3.23.0
ARG TARGETARCH

# Some extra tools we need
RUN apk --no-cache add coreutils procps tini

# rustic
ENV RUSTIC_VERSION=0.10.1
RUN mkdir /tmp/rustic && cd /tmp/rustic  \
    && wget -O rustic.tar.gz https://github.com/rustic-rs/rustic/releases/download/v${RUSTIC_VERSION}/rustic-v${RUSTIC_VERSION}-$(uname -m)-unknown-linux-musl.tar.gz \
    && tar xzf rustic.tar.gz \
    && cp rustic /usr/bin/ \
    && cd / && rm -rf /tmp/rustic

# extra tools we need for dboxed volumes
RUN apk --no-cache add mount umount losetup lvm2 lvm2-dmeventd device-mapper thin-provisioning-tools e2fsprogs btrfs-progs xfsprogs

COPY infra-images/volume/lvm.conf /etc/lvm/lvm.conf
