FROM golang:1.21-alpine AS builder

WORKDIR /build
COPY main.go .
RUN go build -o app main.go

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /build/app .
ENV VERSION=unknown
EXPOSE 8080
CMD ["./app"]
