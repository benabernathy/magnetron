FROM golang:1.20 AS builder

WORKDIR /app
COPY . .

RUN cd cmd/magnetron && go build -o /app/magnetron . && chmod a+x /app/magnetron
RUN /app/magnetron c init /app/config.yml

FROM scratch

COPY --from=builder /app/magnetron /app/magnetron
COPY --from=builder /app/config.yml /usr/local/var/magnetron/config.yml

EXPOSE 5499 5498

CMD ["/app/magnetron", "serve", "/usr/local/var/magnetron/config.yml"]
