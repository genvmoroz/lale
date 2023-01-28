FROM golang:1.19.5-buster as builder

WORKDIR /app

COPY .. .

RUN apt update -y ; \
    apt install -y apt-utils

RUN make ci ; \
    make build

FROM debian:buster

COPY --from=builder /app/bin/svc /svc

EXPOSE 8080

CMD ["./svc"]
