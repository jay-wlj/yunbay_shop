FROM jayden21/alpine:latest

MAINTAINER ybpay "305898636@qq.com"

WORKDIR $GOPATH/src/ybpay


ADD ybpay .

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2


EXPOSE 94


ENTRYPOINT ["./ybpay"]
