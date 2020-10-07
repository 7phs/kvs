FROM golang:1.15-buster

ENV SRC=/go/src/github.com/7phs/kvs

ADD . ${SRC}
WORKDIR ${SRC}

RUN make build

FROM debian:stretch

ENV SRC=/go/src/github.com/7phs/kvs

RUN apt-get update \
    && apt-get install -y ca-certificates \
    && apt-get clean

EXPOSE 9889

WORKDIR /root/
COPY --from=0 ${SRC}/bin ./app

CMD ["./app/kvs"]