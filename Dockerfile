FROM alpine:3.7

MAINTAINER Matt Farmer <matt@frmr.me>

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

RUN wget -O /usr/local/bin/dumb-init https://github.com/Yelp/dumb-init/releases/download/v1.2.1/dumb-init_1.2.1_amd64
RUN chmod +x /usr/local/bin/dumb-init

ADD gifbot /gifbot

ENTRYPOINT ["/usr/local/bin/dumb-init"]

CMD ["/gifbot"]
