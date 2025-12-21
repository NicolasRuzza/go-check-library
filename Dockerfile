# Fase 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY main.go .
# Inicializa e baixa as libs na hora do build
RUN go mod init scraper || true && \
    go get github.com/jomei/notionapi && \
    go get github.com/PuerkitoBio/goquery && \
    go mod tidy
# Compila estático
RUN CGO_ENABLED=0 GOOS=linux go build -o bot .

# Fase 2: Run (Imagem final leve)
FROM alpine:latest
WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/bot .
CMD ["./bot"]