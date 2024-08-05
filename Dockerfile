FROM golang:1.22 AS builder


WORKDIR /app
COPY . .

RUN cd cmd/magnetron && CGO_ENABLED=1 go build -ldflags="-w -s" -o /app/magnetron . && chmod a+x /app/magnetron
RUN /app/magnetron c init /app/config.yml

FROM bitnami/minideb:latest

COPY --from=builder /app/magnetron /app/magnetron
COPY --from=builder /app/config.yml /usr/local/var/magnetron/config.yml

EXPOSE 5499 5498
ENTRYPOINT [ "/app/magnetron" ]
CMD ["serve", "/usr/local/var/magnetron/config.yml"]
