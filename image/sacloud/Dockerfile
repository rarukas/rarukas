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

RUN ["make", "build-server"]

#----------
FROM sacloud/packer as packer-sacloud

FROM hashicorp/packer:light as packer

FROM sacloud/terraform as terraform

FROM sacloud/usacloud as usacloud
#----------

FROM alpine:3.7
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN set -x && apk add --no-cache --update \
       bash \
       curl \
       openssh-client \
       ca-certificates

# Packer
COPY --from=packer /bin/packer /bin/
COPY --from=packer-sacloud /bin/packer-builder-sakuracloud /bin/

# Terraform
COPY --from=terraform /bin/terraform /bin/
COPY --from=terraform /bin/terraform-provider-sakuracloud* /bin/

# Usacloud
COPY --from=usacloud /bin/usacloud /bin/

COPY --from=builder /go/src/github.com/rarukas/rarukas/bin/rarukas-server /usr/local/bin/
RUN chmod +x /usr/local/bin/rarukas-server
ENTRYPOINT ["/usr/local/bin/rarukas-server"]

RUN mkdir /workdir
WORKDIR /workdir

EXPOSE 8080
EXPOSE 2222
