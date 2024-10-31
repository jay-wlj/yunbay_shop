FROM jayden21/alpine:latest

MAINTAINER ybapi "305898636@qq.com"

WORKDIR $GOPATH/src/ybapi



ADD ybapi .

# #设置时区
# RUN apk add tzdata && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
#     && echo 'Asia/Shanghai' > /etc/timezone \
#     && apk del tzdata

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE 90


ENTRYPOINT ["./ybapi"]
