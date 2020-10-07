FROM golang:1.15-buster

ENV SRC=/go/src/github.com/7phs/kvs/test/integrations

ADD . ${SRC}
WORKDIR ${SRC}

CMD ["go", "test", "-v", "-race", "./tst/..."]
