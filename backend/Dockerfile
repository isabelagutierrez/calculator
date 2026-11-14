FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY . ./

RUN go build -o /out/calculator-api ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=builder /out/calculator-api /usr/local/bin/calculator-api

USER appuser
EXPOSE 8080
HEALTHCHECK --interval=10s --timeout=3s --retries=5 CMD wget -qO- http://127.0.0.1:8080/health || exit 1
ENTRYPOINT ["/usr/local/bin/calculator-api"]
