FROM golang:1.25-alpine AS build-base

WORKDIR /src

COPY ./src/go.mod ./src/go.sum ./
RUN go mod download

FROM build-base AS build

ENV GOOS=linux GOARCH=amd64

COPY ./src .
RUN go build -o /build/elaina

FROM debian:stable-slim AS prod-base
LABEL authors="Favouriteless"

RUN apt update
RUN apt install -y ca-certificates

FROM prod-base AS prod

WORKDIR /run
COPY --from=build /build/elaina /run/elaina

ENTRYPOINT ["/run/elaina"]
CMD ["--mode=bot"]