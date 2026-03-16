FROM alpine:3.21

LABEL maintainer="Dmitry Mozzherin"

RUN adduser -D -H gnames

WORKDIR /bin

COPY ./bin/gnames /bin/gnames

USER gnames

ENTRYPOINT [ "gnames" ]

CMD ["rest", "-p", "8888"]
