FROM golang:1.10 AS builder
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN  apt-get update && apt-get -y install \
        bash \
        git  \
        make \
        zip  \
      && apt-get clean \
      && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*

ADD . /go/src/github.com/rarukas/rarukas
WORKDIR /go/src/github.com/rarukas/rarukas

RUN ["make", "build"]

#----------

FROM alpine:3.7
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN set -x && apk add --no-cache --update ca-certificates
COPY --from=builder /go/src/github.com/rarukas/rarukas/bin/rarukas /usr/local/bin/
RUN chmod +x /usr/local/bin/rarukas
ENTRYPOINT ["/usr/local/bin/rarukas"]
