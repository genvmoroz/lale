FROM debian:stable-slim

ARG ARTIFACT

COPY ${ARTIFACT}-linux-amd64 /svc

RUN apt update && apt install -y ca-certificates && rm -rf /var/lib/apt/lists/*

CMD ["./svc"]
