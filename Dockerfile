FROM golang:1.15.10-buster AS builder

RUN mkdir -p /opt/src
COPY cmd /opt/src/cmd
COPY go.* /opt/src/
WORKDIR /opt/src
RUN go build cmd/gifbot.go

FROM debian:buster-slim AS runner

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
COPY --from=builder /opt/src/gifbot /usr/local/bin/gifbot

ENTRYPOINT ["/usr/local/bin/gifbot"]
