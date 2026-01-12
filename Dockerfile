# Build stage
ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm AS builder

WORKDIR /usr/src/app

COPY go.mod ./
COPY go.sum* ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -v -o /run-app ./cmd/server

# Runtime stage
FROM debian:bookworm

COPY --from=builder /run-app /usr/local/bin/
COPY --from=builder /usr/src/app/templates /templates
COPY --from=builder /usr/src/app/static /static

WORKDIR /

EXPOSE 8080

CMD ["run-app"]

