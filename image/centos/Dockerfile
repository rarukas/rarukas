FROM centos:7
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN yum update -y && yum install -y \
        curl \
        git \
        zip \
    && rm -rf /var/cache/yum/* \
    && yum clean all

ADD https://github.com/rarukas/rarukas/releases/download/0.1.0/rarukas-server /usr/local/bin/
RUN chmod +x /usr/local/bin/rarukas-server
ENTRYPOINT ["/usr/local/bin/rarukas-server"]

RUN mkdir /workdir
WORKDIR /workdir

EXPOSE 8080
EXPOSE 2222
