FROM alpine:3.22.1

RUN apk add --no-cache kmod iproute2 nftables

VOLUME /var/lib/dboxed

COPY bin/dboxed /usr/bin/dboxed

ENTRYPOINT ["dboxed"]
