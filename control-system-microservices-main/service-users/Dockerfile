FROM golang:1.25rc3-alpine3.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./cmd/main ./cmd/main.go

FROM alpine:3.22.1

WORKDIR /root/

COPY --from=builder /app/cmd/main .
EXPOSE 8082

CMD ["./main"]