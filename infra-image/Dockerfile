# syntax=docker/dockerfile:1-labs

# dboxed-volume
FROM ghcr.io/dboxed/dboxed-volume:nightly AS dboxed-volume

# rootfs
FROM docker:28.4.0-dind AS rootfs

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

# extra tools we need for deboxed-volume
RUN apk --no-cache add losetup lvm2 device-mapper

# cleanup
RUN apk del git openssh-client

COPY --from=dboxed-volume /usr/bin/rustic /usr/bin/
COPY --from=dboxed-volume /usr/bin/dboxed-volume /usr/bin/

FROM scratch
COPY --from=rootfs / /

COPY services.d /etc/services.d

ENV S6_LOGGING_SCRIPT="n10 s10000000"
ENTRYPOINT ["/init"]
