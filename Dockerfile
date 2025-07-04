ARG WOLFI_DIGEST

FROM golang:1.24.3 AS builder

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETARCH
ENV ARCH=$TARGETARCH

ARG VERSION

ADD . .

ENV GOCACHE=/root/.cache/go-build

RUN --mount=type=cache,target=/go/pkg/mod/ \
    --mount=type=cache,target="/root/.cache/go-build" \
    mv build/vendor vendor && \
    CGO_ENABLED=1 go build -ldflags="-extldflags=-static -X 'main.Version=$VERSION'" -o dist/unboxed ./cmd/unboxed

#####
FROM --platform=$TARGETPLATFORM cgr.dev/chainguard/wolfi-base@$WOLFI_DIGEST

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETPLATFORM

RUN apk add --no-cache kmod iproute2 nftables
RUN for i in iptables iptables-save iptables-restore; do ln -f -s /sbin/xtables-nft-multi /sbin/$i; done

COPY --from=builder /go/dist/unboxed /usr/bin/unboxed

VOLUME /var/lib/koobox

ENTRYPOINT ["unboxed"]
