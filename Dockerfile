FROM golang:1.25-alpine AS dev

# Instala certificados CA (necessario para fazer requisicoes HTTPS para o Notion/Sites)
# Instala o chromium para rodar os sites dinamicos
RUN apk add --no-cache \
    ca-certificates \
    chromium \
    ttf-freefont \
    nss \
    freetype \
    harfbuzz

# As variaveis de ambiente ficam aqui, onde o programa vai rodar
ENV CHROME_BIN=/usr/bin/chromium-browser

WORKDIR /app

COPY go.mod go.sum* ./
RUN if [ ! -f go.mod ]; then go mod init go-check-library; fi
RUN go mod download

# Se no compose houver "target:dev", o Dockerfile sera lido somente ate o cmd, senao o mesmo sera ignorado
CMD ["go", "run", "./cmd/library-scraper/main.go"]


# Continuacao

# Compila
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o go-check-library ./cmd/library-scraper/main.go

FROM alpine:3.23

RUN apk add --no-cache \
    chromium \
    ca-certificates \
    ttf-freefont \
    udev

ENV CHROME_BIN=/usr/bin/chromium-browser

WORKDIR /app-run

COPY --from=dev /app/go-check-library .
CMD ["./go-check-library"]