FROM alpine:3.6

RUN adduser -D elasticsearch-operator
USER elasticsearch-operator

ADD tmp/_output/bin/elasticsearch-operator /usr/local/bin/elasticsearch-operator
