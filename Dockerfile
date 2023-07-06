FROM golang:1.19-bullseye

ARG GO_OS=darwin
ARG GO_ARCH=arm64

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid 65532 \
  small-user

WORKDIR $GOPATH/src/magnetron

COPY . .
RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=${GO_OS} GOARCH=${GO_ARCH} go build -a -installsuffix cgo -o /go/bin/magnetron cmd/magnetron/main.go


RUN mkdir -p /etc/magnetron

RUN /go/bin/magnetron config init /etc/magnetron/config.yml

USER small-user:small-user

CMD ["/go/bin/magnetron serve --config /etc/magnetron/config.yml"]