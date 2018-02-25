FROM golang:1.10 AS builder
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN  apt-get update && apt-get -y install git \
      && apt-get clean \
      && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*

ADD . /go/src/github.com/rarukas/rarukas
WORKDIR /go/src/github.com/rarukas/rarukas

RUN CGO_ENABLED=0 go build \
  -ldflags "-s -w -X github.com/rarukas/rarukas/version.Revision=`git rev-parse --short HEAD`" \
  -o bin/rarukas-server cmd/rarukas-server/main.go

#----------

FROM centos:7
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN yum update -y && yum install -y \
        curl \
        git \
        zip \
    && rm -rf /var/cache/yum/* \
    && yum clean all

COPY --from=builder /go/src/github.com/rarukas/rarukas/bin/rarukas-server /usr/local/bin/
RUN chmod +x /usr/local/bin/rarukas-server
ENTRYPOINT ["/usr/local/bin/rarukas-server"]

RUN mkdir /workdir
WORKDIR /workdir

EXPOSE 8080
EXPOSE 2222
