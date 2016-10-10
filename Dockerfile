FROM dingotiles/dingo-postgresql96-agent-base:latest

# Set up GOPATH
ADD . /go/src/github.com/dingotiles/dingo-postgresql-agent
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN set -x \
    && apk add --update go git \
    && go install github.com/dingotiles/dingo-postgresql-agent \
    && apk del go git \
    && rm -rf /var/cache/apk/*
