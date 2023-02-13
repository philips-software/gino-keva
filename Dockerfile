FROM alpine:3.17.2

RUN \
  apk update && \
  apk add --no-cache git=2.32.0-r0 && \
  rm /var/cache/apk/*

COPY build/gino-keva /usr/local/bin/
CMD [ "gino-keva", "list" ]
