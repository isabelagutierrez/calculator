FROM golang:1.27-alpine AS builder

WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . ./

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/calculator-api ./cmd/server

FROM alpine:3.22

RUN adduser -D -H -u 10001 appuser
COPY --from=builder /out/calculator-api /usr/local/bin/calculator-api

USER appuser
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/calculator-api"]
