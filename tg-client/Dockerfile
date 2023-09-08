FROM golang:1.21 as builder

WORKDIR /app

COPY .. .

RUN apt update -y ; \
    apt install -y apt-utils

RUN make build

FROM debian:stable-slim

RUN apt update && apt install -y apt-utils ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/svc /svc

CMD ["./svc"]
