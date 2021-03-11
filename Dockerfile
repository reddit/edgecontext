FROM ghcr.io/reddit/thrift-compiler:0.14.1 AS thrift
FROM docker.io/library/golang:1.16.1-buster AS go
FROM python:3.8

COPY --from=thrift /usr/local/bin/thrift /usr/local/bin/thrift
COPY --from=go /usr/local/go /usr/local/go

ENV GOPATH=/root/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH
RUN go install golang.org/x/lint/golint@latest

WORKDIR /src
COPY lib/py/requirements*.txt ./
RUN pip install -r requirements.txt
CMD ["/bin/bash"]
