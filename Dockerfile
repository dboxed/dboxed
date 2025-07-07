# use `docker buildx imagetools inspect cgr.dev/chainguard/wolfi-base:latest` to find latest sha256 of multiarch image
FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/wolfi-base@sha256:1c6a85817d3a8787e094aae474e978d4ecdf634fd65e77ab28ffae513e35cca1

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETPLATFORM

RUN apk add --no-cache kmod iproute2 nftables
RUN for i in iptables iptables-save iptables-restore; do ln -f -s /sbin/xtables-nft-multi /sbin/$i; done

VOLUME /var/lib/unboxed

COPY bin/unboxed /usr/bin/unboxed

ENTRYPOINT ["unboxed"]
