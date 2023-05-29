FROM golang:1.20.4 as builder

WORKDIR /app

COPY .. .

RUN apt update -y ; \
    apt install -y apt-utils

RUN make ci ; \
    make build

FROM debian:buster-slim

RUN apt update && apt install -y apt-utils ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/svc /svc

CMD ["./svc"]
