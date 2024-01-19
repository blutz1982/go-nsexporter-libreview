FROM golang:1.21.4 as build

ENV GO111MODULE=on

RUN mkdir /opt/build && go env

WORKDIR /opt/build

COPY . ./

RUN make build

FROM ubuntu:focal

RUN apt update && \
    DEBIAN_FRONTEND=noninteractive TZ=Europe/Moscow apt install -y ca-certificates tzdata

ENV WORKDIR=/usr/local/bin/

COPY --from=build /opt/build/build/nsexport ${WORKDIR}

WORKDIR ${WORKDIR}

ENTRYPOINT ["/usr/local/bin/nsexport"]
