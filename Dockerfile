FROM alpine:3.17

LABEL maintainer="Dmitry Mozzherin"

WORKDIR /bin

COPY ./gnames /bin

ENTRYPOINT [ "gnames" ]

CMD ["rest", "-p", "8888"]
