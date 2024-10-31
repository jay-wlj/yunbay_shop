FROM jayden21/alpine:latest

MAINTAINER ybcron "305898636@qq.com"

WORKDIR $GOPATH/src/ybcron


ADD ybcron .

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2


ENTRYPOINT ["./ybcron"]

