# =====================================================
# VIP Lounge - Dockerfile para Deploy no Render
# =====================================================

# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Instalar dependências de build
RUN apk add --no-cache git ca-certificates tzdata

# Definir diretório de trabalho
WORKDIR /app

# Copiar arquivos de dependências primeiro (para cache de layers)
COPY go.mod go.sum ./

# Baixar dependências
RUN go mod download

# Copiar código fonte
COPY . .

# Compilar o binário
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/server

# Stage 2: Runtime
FROM alpine:3.19

# Instalar certificados SSL e timezone data
RUN apk --no-cache add ca-certificates tzdata

# Criar usuário não-root para segurança
RUN adduser -D -g '' appuser

# Definir diretório de trabalho
WORKDIR /app

# Copiar binário compilado
COPY --from=builder /app/server .

# Copiar arquivos estáticos do frontend (se existirem)
COPY --from=builder /app/web ./web

# Copiar config.yaml (opcional, pode ser sobrescrito por env vars)
COPY --from=builder /app/config.yaml ./config.yaml

# Mudar para usuário não-root
USER appuser

# =====================================================
# IMPORTANTE PARA O RENDER:
# O Render define a variável $PORT automaticamente
# O servidor DEVE escutar nessa porta
# =====================================================
EXPOSE ${PORT:-8080}

# Healthcheck para o Render
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# Comando de inicialização
CMD ["./server"]
