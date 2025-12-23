# syntax=docker/dockerfile:1-labs

# rootfs
FROM alpine:3.23.0
ARG TARGETARCH

# Some extra tools we need
RUN apk --no-cache add coreutils procps tini

# extra tools we need for dboxed volumes
RUN apk --no-cache add mount umount losetup lvm2 lvm2-dmeventd device-mapper thin-provisioning-tools e2fsprogs btrfs-progs xfsprogs

# restic
ENV RESTIC_VERSION=0.18.1
RUN mkdir /tmp/restic && cd /tmp/restic  \
    && wget -O restic.bz2 https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}/restic_${RESTIC_VERSION}_linux_${TARGETARCH}.bz2 \
    && bunzip2 restic.bz2 \
    && chmod +x restic \
    && mv restic /usr/bin/ \
    && cd / && rm -rf /tmp/restic

COPY infra-images/volume/lvm.conf /etc/lvm/lvm.conf
