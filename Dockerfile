FROM golang:alpine AS dev

# Instala certificados CA (necessario para fazer requisicoes HTTPS para o Notion/Sites)
# Instala o chromium para rodar os sites dinamicos
RUN apk add --no-cache \
    ca-certificates \
    chromium \
    ttf-freefont

# As variaveis de ambiente ficam aqui, onde o programa vai rodar
ENV CHROME_BIN=/usr/bin/chromium-browser

WORKDIR /app

COPY go.mod go.sum* ./
RUN if [ ! -f go.mod ]; then go mod init go-check-library; fi
RUN go mod tidy

# Se no compose houver "target:dev", o Dockerfile sera lido somente ate o cmd, senao o mesmo sera ignorado
CMD ["go", "run", "main.go"]


# Continuacao

# Compila
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o go-check-library main.go

FROM alpine:latest

RUN apk add --no-cache chromium ca-certificates ttf-freefont

ENV CHROME_BIN=/usr/bin/chromium-browser

WORKDIR /app-run

COPY --from=builder /app/go-check-library .
CMD ["./go-check-library"]