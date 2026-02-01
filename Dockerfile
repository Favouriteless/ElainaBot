FROM golang:1.25-alpine AS build-base

WORKDIR /src

COPY ./src/elainabot/go.mod ./src/elainabot/go.sum ./
RUN go mod download

FROM build-base AS build

ENV GOOS=linux GOARCH=amd64

COPY ./src .
RUN go build -C ./elainabot -o /build/elaina

FROM debian:stable-slim AS prod-base
LABEL authors="Favouriteless"

RUN apt update
RUN apt install -y ca-certificates

FROM prod-base AS prod

WORKDIR /run
COPY ./src/common/migrations /run/migrations
COPY --from=build /build/elaina /run/elaina

ENTRYPOINT ["/run/elaina"]
CMD ["--mode=bot"]