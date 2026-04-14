FROM golang:1.25-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux go build -o /out/server ./cmd/server

FROM alpine:3.20

RUN adduser -D appuser
WORKDIR /app

COPY --from=builder /out/server /app/server

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/server"]
