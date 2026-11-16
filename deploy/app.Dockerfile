FROM golang:1.27.3-alpine AS build
WORKDIR /src

RUN apk add --no-cache git curl bash
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY . .
RUN task build

###
FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 65532 app

COPY --from=build --chown=app:app /src/bin/notify /app/notify

USER app

EXPOSE 8080 9090

ENTRYPOINT ["/app/notify"]
