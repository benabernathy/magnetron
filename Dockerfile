FROM golang:1.25.5 AS builder


WORKDIR /app
COPY . .

RUN cd cmd/magnetron && CGO_ENABLED=1 go build -ldflags="-w -s" -o /app/magnetron . && chmod a+x /app/magnetron
RUN /app/magnetron c init /app/config.yml


FROM gcr.io/distroless/base-debian10

COPY --from=builder /app/magnetron /app/magnetron
#COPY --from=builder /app/config.yml /usr/local/var/magnetron/config.yml

EXPOSE 5499 5498 8080
#ENTRYPOINT [ "/app/magnetron" ]
CMD ["/app/magnetron", "serve", "/usr/local/var/magnetron/config.yml"]
