FROM dingotiles/dingo-postgresql96-agent-base:latest

ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN apk add --no-cache go git

RUN set -x \
      && echo "@edge http://nl.alpinelinux.org/alpine/edge/main" >> /etc/apk/repositories \
      && apk add --update pstree@edge

COPY config/patroni-default-values.yml /patroni/patroni-default-values.yml
COPY images/scripts/* /scripts/
COPY images/motd /etc/motd
RUN echo "source /etc/motd" >> /root/.bashrc
COPY images/supervisord.conf /etc/supervisor/supervisord.conf
COPY images/services/*.conf /etc/supervisor/conf.d/

COPY . /go/src/github.com/dingotiles/dingo-postgresql-agent
RUN set -x \
    && go install github.com/dingotiles/dingo-postgresql-agent \
    && rm -rf $GOPATH/src
