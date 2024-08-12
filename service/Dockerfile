FROM golang:1.22 AS builder

WORKDIR /app

COPY . .

RUN apt update -y ; \
    apt install -y apt-utils

RUN make build

FROM debian:stable-slim

RUN apt update && apt install -y apt-utils ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/svc /svc

EXPOSE 8080

CMD ["./svc"]
