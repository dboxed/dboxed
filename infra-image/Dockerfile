# syntax=docker/dockerfile:1-labs

# rootfs
FROM docker:28.4.0-dind AS rootfs
ARG TARGETARCH

# Some extra tools we need
RUN apk --no-cache add coreutils procps iproute2

ENV S6_OVERLAY_VERSION=3.2.1.0

ADD infra-image/s6-overlay.sha256 /tmp/s6-overlay/
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
    && mkdir -p /var/lib/dboxed && chmod 0700 /var/lib/dboxed \
    && mkdir -p /var/lib/dboxed/logs \
    && mkdir -p /var/lib/dboxed/containers \
    && mkdir -p /run/netns

RUN mkdir /var/lib/docker && ln -s /var/lib/dboxed/containers /var/lib/docker/containers

# Netbird
ENV NETBIRD_VERSION=0.59.12
RUN mkdir /tmp/netbird && cd /tmp/netbird  \
    && wget -O netbird.tar.gz https://github.com/netbirdio/netbird/releases/download/v${NETBIRD_VERSION}/netbird_${NETBIRD_VERSION}_linux_${TARGETARCH}.tar.gz \
    && tar xzf netbird.tar.gz \
    && cp netbird /usr/bin/ \
    && cd / && rm -rf /tmp/netbird

RUN mkdir /var/lib/dboxed/netbird && ln -s /var/lib/dboxed/netbird /var/lib/netbird

# rustic
ENV RUSTIC_VERSION=0.10.1
RUN mkdir /tmp/rustic && cd /tmp/rustic  \
    && wget -O rustic.tar.gz https://github.com/rustic-rs/rustic/releases/download/v${RUSTIC_VERSION}/rustic-v${RUSTIC_VERSION}-$(uname -m)-unknown-linux-musl.tar.gz \
    && tar xzf rustic.tar.gz \
    && cp rustic /usr/bin/ \
    && cd / && rm -rf /tmp/rustic

# extra tools we need for dboxed volumes
RUN apk --no-cache add mount umount losetup lvm2 lvm2-dmeventd device-mapper thin-provisioning-tools

# cleanup
RUN apk --no-cache del git openssh-client

FROM scratch
COPY --from=rootfs / /

COPY infra-image/services.d /etc/services.d
COPY infra-image/lvm.conf /etc/lvm/lvm.conf

RUN mkdir /etc/docker
COPY infra-image/daemon.json /etc/docker/
COPY infra-image/resolv.conf /etc/resolv.conf

ENV S6_LOGGING_SCRIPT="n10 s10000000"
ENV S6_KEEP_ENV=1
ENV DBOXED_SANDBOX=1
ENV PATH=/command:$PATH
ENTRYPOINT ["/init"]
