FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./

RUN go build -o exporter -v ./...

FROM alpine:latest
COPY --from=builder /app/exporter /exporter

EXPOSE 8080

CMD ["/exporter"]
