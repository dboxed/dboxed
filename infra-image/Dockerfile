# syntax=docker/dockerfile:1-labs

FROM docker:28.3.2-dind as initial

RUN apk add --no-cache \
    tini

RUN mkdir -p /etc/dboxed \
    && mkdir -p /var/lib/dboxed \
    && mkdir -p /var/lib/dboxed/logs \
    && mkdir -p /run/netns \
    && mkdir -p /hostfs

# cleanup
RUN apk del git openssh-client e2fsprogs e2fsprogs-extra btrfs-progs xfsprogs zfs

FROM scratch
COPY --from=initial / /
