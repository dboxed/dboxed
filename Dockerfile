FROM alpine:3.22.1
ARG TARGETPLATFORM

RUN apk add --no-cache kmod iproute2 nftables

VOLUME /var/lib/dboxed

COPY $TARGETPLATFORM/dboxed /usr/bin/dboxed

ENTRYPOINT ["dboxed"]
