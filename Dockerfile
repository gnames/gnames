FROM alpine

LABEL maintainer="Dmitry Mozzherin"

WORKDIR /bin

COPY ./gnames/gnames /bin

ENTRYPOINT [ "gnames" ]

CMD ["rest", "-p", "8888"]