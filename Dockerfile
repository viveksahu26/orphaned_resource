FROM docker.io/library/golang:1.18-alpine as buildStage

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=0 go build -v -o obmondo-k8s-agent .

ENV PROMETHEUS_URL http://localhost:9090

ENV API_URL https://obmondo.com

FROM registry.obmondo.com/obmondo/dockerfiles/ubuntu:22.04

RUN apt-get update && apt install -y ca-certificates

WORKDIR /app

COPY --from=buildStage /app/obmondo-k8s-agent /app/

CMD [ "/app/obmondo-k8s-agent" ]
