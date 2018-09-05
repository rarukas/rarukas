FROM debian:stretch
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN  apt-get update && apt-get -y install \
        bash \
        git  \
        curl \
        zip \
      && apt-get clean \
      && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*

ADD https://github.com/rarukas/rarukas/releases/download/0.1.0/rarukas-server /usr/local/bin/
RUN chmod +x /usr/local/bin/rarukas-server
ENTRYPOINT ["/usr/local/bin/rarukas-server"]

RUN mkdir /workdir
WORKDIR /workdir

EXPOSE 8080
EXPOSE 2222
