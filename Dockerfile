FROM ghcr.io/reddit/thrift-compiler:0.14.0 AS thrift
FROM python:3.8

COPY --from=thrift /usr/local/bin/thrift /usr/local/bin/thrift

WORKDIR /src
COPY lib/py/requirements*.txt .
RUN pip install -r requirements.txt
CMD ["/bin/bash"]
