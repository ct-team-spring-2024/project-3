FROM golang:1.23.4-bookworm AS builder
WORKDIR /app
COPY ./go.mod ./go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o /app/out ./loadbalancer/cmd/main.go
FROM alpine:latest AS final
EXPOSE 8081
COPY --from=builder /app/out /app/loadbalancer
WORKDIR /app
CMD ["./loadbalancer"]
