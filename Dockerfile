# Second stage container
FROM alpine:3.16

ENV TZ=Asia/Shanghai
RUN apk add tzdata && cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone

WORKDIR /strategy
COPY ./build/bbgo /usr/local/bin

ENTRYPOINT ["/usr/local/bin/bbgo"]