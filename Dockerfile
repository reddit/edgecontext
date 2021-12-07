FROM ghcr.io/reddit/thrift-compiler:0.14.1 AS thrift
FROM docker.io/library/golang:1.18.2-buster AS go
FROM python:3.9

COPY --from=thrift /usr/local/bin/thrift /usr/local/bin/thrift
COPY --from=go /usr/local/go /usr/local/go

RUN mkdir /opt/go
ENV GOPATH=/opt/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
RUN go install golang.org/x/lint/golint@latest

WORKDIR /src
COPY lib/py/requirements*.txt ./
RUN pip install -r requirements.txt

# Permission fixes for running as a test environment.
# Assumption: This docker image is not used outside build and test
# environments.
# $GOPATH: conflicts with local volume mount of /src:
#    This mount provides a live code editing environment.
#    It is not accessed as root (must use the host's UID namespace to write
#    back to the volume) so to have a single process reading $GOPATH and /src
#    simultaneously we need open permissions.
# $HOME/.cache:
#    go expects $HOME/.cache to be writable but in the recommended test
#    environment, $HOME is /, so pre-create .cache too.

RUN mkdir /.cache
RUN chmod a+rwX -R $GOPATH /.cache

CMD ["/bin/bash"]
