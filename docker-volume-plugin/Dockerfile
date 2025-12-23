FROM alpine:3.22.1
ARG TARGETPLATFORM
ARG TARGETARCH

RUN apk add --no-cache kmod iproute2 nftables

# restic
ENV RESTIC_VERSION=0.18.1
RUN mkdir /tmp/restic && cd /tmp/restic  \
    && wget -O restic.bz2 https://github.com/restic/restic/releases/download/v${RESTIC_VERSION}/restic_${RESTIC_VERSION}_linux_${TARGETARCH}.bz2 \
    && bunzip2 restic.bz2 \
    && chmod +x restic \
    && mv restic /usr/bin/ \
    && cd / && rm -rf /tmp/restic

VOLUME /var/lib/dboxed

RUN mkdir -p /run/docker/plugins 
# RUN mkdir -p /var/lib/dboxed/volumes

COPY $TARGETPLATFORM/dboxed /usr/bin/dboxed

ENTRYPOINT ["dboxed"]
