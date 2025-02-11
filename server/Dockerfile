FROM golang:1.23-alpine AS builder

WORKDIR /src
COPY . .

RUN apk add make && make clean build

FROM cgr.dev/chainguard/wolfi-base

WORKDIR /app

RUN addgroup --system --gid 2001 server && \
    adduser --system --uid 2001 server && \
    mkdir -p /app/dbs && \
    chown -R server:server /app

USER server:server

ENV PWD=/app
COPY --from=builder --chown=server:server /src/out/server .

CMD ["/app/server"]
