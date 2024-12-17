FROM --platform=linux/amd64 debian:stable-slim

ARG ARTIFACT

COPY ${ARTIFACT}-linux-amd64 /svc

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

EXPOSE 8080

CMD ["sh", "-c", "./svc 2>&1 | tee -a /var/log/service.log"]
