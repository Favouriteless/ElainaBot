FROM golang:1.25-alpine AS build-base

WORKDIR /src

# Copy module & workspace definitions
COPY go.work go.work.sum ./
COPY bot/go.mod bot/go.sum bot/
COPY common/go.mod common/go.sum common/

RUN go mod -C ./bot download

FROM build-base AS build-final

ENV GOOS=linux GOARCH=amd64

COPY bot bot
COPY common common

RUN go build -C ./bot -o /build/elaina

FROM debian:stable-slim AS prod-base
LABEL authors="Favouriteless"

RUN apt update
RUN apt install -y ca-certificates

FROM prod-base AS prod

COPY --from=build-final /src/common/migrations /run/migrations
COPY --from=build-final /build/elaina /run/elaina

ENTRYPOINT ["/run/elaina"]
CMD ["--mode=bot"]