# syntax=docker/dockerfile:1-labs

# rootfs
FROM alpine:3.23.0
ARG TARGETARCH

# Some extra tools we need
RUN apk --no-cache add coreutils procps tini

# extra tools we need for dboxed volumes
RUN apk --no-cache add mount umount losetup lvm2 lvm2-dmeventd device-mapper thin-provisioning-tools e2fsprogs btrfs-progs xfsprogs

COPY infra-images/volume/lvm.conf /etc/lvm/lvm.conf
