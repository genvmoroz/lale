FROM golang:1.23 AS builder
#todo: refactor the file to use already built image instead of building from scratch
WORKDIR /app

COPY . .

RUN apt update -y ; \
    apt install -y apt-utils

RUN make build

FROM debian:stable-slim

RUN apt update && apt install -y apt-utils ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/svc /svc

EXPOSE 8080

CMD ["sh", "-c", "./svc 2>&1 | tee -a /var/log/service.log"]
