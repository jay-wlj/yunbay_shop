FROM jayden21/alpine:latest

MAINTAINER yborder "305898636@qq.com"

WORKDIR $GOPATH/src/yborder

ADD conf conf/

ADD yborder .

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN mkdir -p /tmp/logs/yborder

EXPOSE 91


ENTRYPOINT ["./yborder"]
