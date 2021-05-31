FROM alpine:3.13

RUN \
  apk update && \
  apk add --no-cache git=2.30.2-r0 && \
  rm /var/cache/apk/*

COPY build/gino-keva /
ENTRYPOINT [ "/gino-keva" ]
