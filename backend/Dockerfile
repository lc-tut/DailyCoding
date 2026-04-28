# ── Stage 1: Build ──────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app ./cmd/server

# ── Stage 2: Runtime ────────────────────────────────────────────
FROM alpine:3.21

# docker-cli: コード実行エンジンがホスト Docker ソケットを使うために必要
RUN apk add --no-cache docker-cli ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app ./server

# マイグレーションファイルを埋め込む
COPY internal/infrastructure/db/migrations ./migrations

ENV TZ=Asia/Tokyo
EXPOSE 8080
ENTRYPOINT ["/app/server"]
