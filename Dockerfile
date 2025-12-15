FROM haproxy:3.3-alpine

WORKDIR /all

COPY . .

USER root

RUN set -eux; \
    apk add --no-cache --virtual .build-deps make go; \
    make build-release-x64 PROJECT_NAME=roblox-load-balancer-daemon; \
    apk del --no-network .build-deps

USER haproxy

ENTRYPOINT ["/all/bin/release/linux/x64/roblox-load-balancer-daemon"]
