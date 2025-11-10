FROM alpine:3.22.1
ARG TARGETPLATFORM

RUN apk add --no-cache kmod iproute2 nftables

ENV RUSTIC_VERSION=0.10.1
RUN mkdir /tmp/rustic && cd /tmp/rustic  \
    && wget -O rustic.tar.gz https://github.com/rustic-rs/rustic/releases/download/v${RUSTIC_VERSION}/rustic-v${RUSTIC_VERSION}-$(uname -m)-unknown-linux-musl.tar.gz \
    && tar xzf rustic.tar.gz \
    && cp rustic /usr/bin/ \
    && cd / && rm -rf /tmp/rustic

VOLUME /var/lib/dboxed

COPY $TARGETPLATFORM/dboxed /usr/bin/dboxed

ENTRYPOINT ["dboxed"]
