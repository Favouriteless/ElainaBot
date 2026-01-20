FROM debian:stable-slim AS base
LABEL authors="Favouriteless"

WORKDIR /run
COPY ./build/elaina elaina

RUN apt update
RUN apt install -y ca-certificates

ENTRYPOINT ["/run/elaina"]
CMD ["--mode=bot"]