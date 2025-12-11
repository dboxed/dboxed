# syntax=docker/dockerfile:1-labs

# rootfs
FROM docker:28.4.0-dind AS rootfs
ARG TARGETARCH

# Some extra tools we need
RUN apk --no-cache add coreutils procps iproute2 s6 tini bind-tools
RUN apk --no-cache add mount umount losetup lvm2

RUN mkdir -p /etc/dboxed \
    && mkdir -p /var/lib/dboxed && chmod 0700 /var/lib/dboxed \
    && mkdir -p /var/lib/dboxed/logs \
    && mkdir -p /run/netns

# cleanup
RUN apk --no-cache del git openssh-client

# busybox image
FROM alpine:3.23.0 AS busybox-image
ENV BUSYBOX_IMAGE_VERSION=1.37.0-musl
RUN apk add skopeo
RUN skopeo copy docker://busybox:$BUSYBOX_IMAGE_VERSION docker-archive:/busybox-image.tar:dboxed-busybox:latest

FROM scratch
COPY --from=rootfs / /
COPY --from=busybox-image /busybox-image.tar /

RUN mkdir /etc/docker
COPY infra-images/sandbox/daemon.json /etc/docker/
COPY infra-images/sandbox/resolv.conf /etc/resolv.conf

COPY infra-images/sandbox/sandbox-entrypoint.sh /sandbox-entrypoint.sh

ENTRYPOINT ["tini", "-g", "/sandbox-entrypoint.sh"]
