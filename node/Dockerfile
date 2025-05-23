FROM golang:1.23.4-bookworm AS builder
WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o /app/out ./node/cmd/main.go
FROM alpine:latest AS final
COPY --from=builder /app/out /app/node
WORKDIR /app
CMD ["./node"]
