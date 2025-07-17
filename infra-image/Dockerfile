# syntax=docker/dockerfile:1-labs

FROM docker:28.3.2-dind as initial

RUN apk add --no-cache \
    tini

RUN mkdir -p /etc/unboxed \
    && mkdir -p /var/log/unboxed \
    && mkdir -p /run/netns \
    && mkdir -p /hostfs

# cleanup
RUN apk del git openssh-client e2fsprogs e2fsprogs-extra btrfs-progs xfsprogs zfs
RUN rm /usr/local/libexec/docker/cli-plugins/docker-buildx

FROM scratch
COPY --from=initial / /
