FROM alpine:3.14.0

RUN \
  apk update && \
  apk add --no-cache git=2.30.2-r0 && \
  rm /var/cache/apk/*

COPY build/gino-keva /usr/local/bin/
CMD [ "gino-keva", "list" ]
