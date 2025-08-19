# syntax=docker/dockerfile:1-labs

FROM docker:28.3.2-dind AS initial

ENV S6_OVERLAY_VERSION=3.2.1.0

ADD s6-overlay.sha256 /tmp/s6-overlay/
RUN cd /tmp/s6-overlay \
    && wget https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-noarch.tar.xz \
    && wget https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-aarch64.tar.xz \
    && wget https://github.com/just-containers/s6-overlay/releases/download/v${S6_OVERLAY_VERSION}/s6-overlay-x86_64.tar.xz \
    && sha256sum *.tar.xz > s6-overlay.sha256.new \
    && diff -u s6-overlay.sha256 s6-overlay.sha256.new

RUN cd /tmp/s6-overlay \
    && tar -C / -Jxpf s6-overlay-noarch.tar.xz \
    && tar -C / -Jxpf s6-overlay-$(uname -m).tar.xz

RUN rm -rf /tmp/s6-overlay

RUN mkdir -p /etc/dboxed \
    && mkdir -p /var/lib/dboxed \
    && mkdir -p /var/lib/dboxed/logs \
    && mkdir -p /run/netns \
    && mkdir -p /hostfs

# cleanup
RUN apk del git openssh-client e2fsprogs e2fsprogs-extra btrfs-progs xfsprogs zfs

FROM scratch
COPY --from=initial / /

COPY services.d /etc/services.d

ENV S6_LOGGING_SCRIPT="n10 s10000000"
ENTRYPOINT ["/init"]
