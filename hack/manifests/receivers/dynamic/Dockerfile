FROM golang:1.25 AS builder

WORKDIR /build

COPY main.go .

RUN CGO_ENABLED=0 go build -o app main.go

FROM registry.access.redhat.com/ubi9/ubi:latest

WORKDIR /app

COPY --from=builder /build/app .

EXPOSE 8080

CMD ["/app/app"]
