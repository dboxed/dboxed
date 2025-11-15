ARG CADDY_VERSION

FROM caddy:${CADDY_VERSION}-builder AS builder

RUN go env -w GOCACHE=/go-cache
RUN go env -w GOMODCACHE=/gomod-cache

ADD caddy/certmagic-dboxed certmagic-dboxed
RUN --mount=type=cache,target=/gomod-cache --mount=type=cache,target=/go-cache \
    xcaddy build --with github.com/dboxed/dboxed-caddy-certmagic=./certmagic-dboxed

FROM caddy:${CADDY_VERSION}-alpine
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
