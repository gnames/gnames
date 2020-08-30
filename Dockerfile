FROM alpine

LABEL maintainer="Dmitry Mozzherin"

# RUN apk add --no-cache bash

WORKDIR /bin

COPY ./gnames/gnames /bin

ENTRYPOINT [ "gnames" ]

CMD ["rest", "-p", "8080"]
