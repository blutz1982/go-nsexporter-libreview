FROM golang:1.19.0-buster as build

ENV GO111MODULE=on

RUN mkdir /opt/build && go env

WORKDIR /opt/build

COPY . ./

RUN make build

FROM ubuntu:xenial

RUN apt update && \
    apt install -y libssl1.0.0 ca-certificates

ENV WORKDIR=/usr/local/bin/

COPY --from=build /opt/build/nsexport ${WORKDIR}

WORKDIR ${WORKDIR}

ENTRYPOINT ["/usr/local/bin/nsexport"]
