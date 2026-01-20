# Script para iniciar o backend em localhost
# Uso: .\start-backend.ps1

Write-Host "╔═══════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║   🚀 Iniciando Backend VIP Lounge         ║" -ForegroundColor Green
Write-Host "╚═══════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

# Definir porta
$env:PORT = "8081"

Write-Host "📌 Porta: 8081" -ForegroundColor Cyan
Write-Host "📂 Diretório: $PWD" -ForegroundColor Cyan
Write-Host ""
Write-Host "⏳ Compilando e iniciando servidor..." -ForegroundColor Yellow
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
Write-Host ""

# Rodar servidor
go run cmd/server/main.go

# Se chegar aqui, servidor parou
Write-Host ""
Write-Host "❌ Servidor parou!" -ForegroundColor Red
Read-Host "Pressione Enter para fechar"
