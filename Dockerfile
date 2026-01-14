FROM golang:1.26-rc-alpine AS builder

WORKDIR /app

COPY . .

RUN if [ ! -f go.mod ]; then go mod init go-check-library; fi

RUN go mod tidy

# Compila
RUN CGO_ENABLED=0 GOOS=linux go build -o go-check-library main.go

FROM alpine:latest

# Instala certificados CA (necessário para fazer requisições HTTPS para o Notion/Sites)
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/go-check-library .

CMD ["./go-check-library"]