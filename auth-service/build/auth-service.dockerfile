FROM golang:1.23.2-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o authApp ./cmd/app/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/authApp /app/
COPY --from=builder /app/configs/*.env /app/configs/

CMD ["/app/authApp"]