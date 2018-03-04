FROM debian:jessie-slim

MAINTAINER Matt Farmer <matt@frmr.me>

ADD https://github.com/Yelp/dumb-init/releases/download/v1.2.1/dumb-init_1.2.1_amd64 /usr/local/bin/dumb-init

RUN chmod +x /usr/local/bin/dumb-init

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

ADD gifbot /usr/local/bin/gifbot

ENTRYPOINT ["/usr/local/bin/dumb-init"]

CMD ["/usr/local/bin/gifbot"]
