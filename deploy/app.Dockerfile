FROM golang:1.27.3-alpine AS build
WORKDIR /src

RUN apk add --no-cache git curl bash
RUN sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

COPY . .
RUN task build
RUN task goose:install

###
FROM alpine:3.21
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -H -u 65532 app

COPY --from=build /src/bin/goose /app/bin/goose
COPY --from=build /usr/local/bin/task /usr/local/bin/task

COPY --from=build --chown=app:app /src/bin/notify /app/notify
COPY --from=build /src/Taskfile.yaml /app/Taskfile.yaml
COPY --from=build /src/migration/ /app/migration/

COPY --chown=app:app migration/ /app/migration/
RUN chmod 755 /app/bin/goose

USER app

EXPOSE 8080 9090

# ENTRYPOINT ["task", "goose:up-all"]
# CMD ["/app/notify"]
ENTRYPOINT ["/bin/sh", "-c", "task goose:up-all && exec /app/notify"]