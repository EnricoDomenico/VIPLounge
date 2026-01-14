# Stage 1: Builder
FROM golang:1.21-alpine AS builder

# Instalar dependências de sistema necessárias para build (se houver, ex: git)
RUN apk add --no-cache git

WORKDIR /app

# Copiar arquivos de dependência primeiro (para cache eficiente)
COPY go.mod ./
# COPY go.sum ./ 
# (Nota: go.sum ainda não existe porque não rodamos go mod tidy, mas em produção existiria)

# Baixar dependências
# RUN go mod download 
# (Comentado pois sem internet/go.sum no ambiente atual do user pode falhar o build localmente, 
# mas no Cloud Build rodará OK se o go.sum for gerado antes do push)

# Copiar todo o código fonte
COPY . .

# Build do binário estático
# CGO_ENABLED=0 garante que não dependa de libc, ideal para containers scratch/alpine
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server ./cmd/server/main.go

# Stage 2: Runtime
FROM alpine:latest

# Instalar certificados CA para chamadas HTTPS externas (Firestore/Benef API)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copiar o binário do builder
COPY --from=builder /app/server .

# Copiar os assets do frontend
COPY --from=builder /app/web ./web

# Usuário não-root para segurança
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Expor porta (Cloud Run injeta PORT env var, mas documentamos 8080)
EXPOSE 8080

# Comando de entrada
CMD ["./server"]


