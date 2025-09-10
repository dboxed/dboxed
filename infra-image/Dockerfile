# syntax=docker/dockerfile:1-labs

# rustic
FROM alpine:3.22 AS rustic
RUN apk add git

RUN mkdir /tools

ENV RUSTIC_NIGHTLY_REVISION=edb3827e39b4c242463bb322829e1afbb67a74dc
RUN cd /tmp && git clone https://github.com/rustic-rs/nightly.git --revision=$RUSTIC_NIGHTLY_REVISION --depth=1
RUN cd /tmp/nightly/rustic/ && tar xzf rustic-nightly-$(uname -m)-unknown-linux-gnu.tar.gz
RUN mv /tmp/nightly/rustic/rustic /tools/
RUN rm -rf /tmp/nightly

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
RUN apk --no-cache add losetup lvm2

# cleanup
RUN apk del git openssh-client

COPY --from=rustic /tools/* /usr/bin/

FROM scratch
COPY --from=rootfs / /

COPY services.d /etc/services.d

ENV S6_LOGGING_SCRIPT="n10 s10000000"
ENTRYPOINT ["/init"]
