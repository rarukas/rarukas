FROM alpine:3.7
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN apk --update add sudo python3 openssl ca-certificates bash curl openssh-client && \
    python3 -m ensurepip && \
    rm -r /usr/lib/python*/ensurepip && \
    pip3 install --upgrade pip setuptools && \
    if [ ! -e /usr/bin/pip ]; then ln -s pip3 /usr/bin/pip ; fi && \
    if [[ ! -e /usr/bin/python ]]; then ln -sf /usr/bin/python3 /usr/bin/python; fi && \
    rm -r /root/.cache && \
    apk --update add --virtual build-dependencies python3-dev libffi-dev openssl-dev build-base  && \
    pip install --upgrade pip cffi && \
    pip install ansible ansible-cmdb && \
    apk del build-dependencies            && \
    rm -rf /var/cache/apk/*               && \
    mkdir -p /etc/ansible                        && \
    echo 'localhost' > /etc/ansible/hosts

ADD https://github.com/rarukas/rarukas/releases/download/0.2.0/rarukas-server /usr/local/bin/
RUN chmod +x /usr/local/bin/rarukas-server
ENTRYPOINT ["/usr/local/bin/rarukas-server"]

RUN mkdir /workdir
WORKDIR /workdir

EXPOSE 8080
EXPOSE 2222
