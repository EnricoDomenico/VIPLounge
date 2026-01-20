# Script para testar o backend
# Uso: .\test-backend.ps1

Write-Host "╔═══════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║   🧪 Testando Backend VIP Lounge          ║" -ForegroundColor Green
Write-Host "╚═══════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

$baseUrl = "http://localhost:8081"

# Testar Health
Write-Host "📡 Testando: GET $baseUrl/v1/health" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "$baseUrl/v1/health" -UseBasicParsing
    Write-Host "✅ Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "📄 Resposta: $($response.Content)" -ForegroundColor Green
} catch {
    Write-Host "❌ Erro: $_" -ForegroundColor Red
}

Write-Host ""

# Testar Validate
Write-Host "📡 Testando: POST $baseUrl/v1/validate" -ForegroundColor Cyan
try {
    $body = @{
        cpf = "123.456.789-00"
        condo_id = "condo_demo_123"
    } | ConvertTo-Json

    $response = Invoke-WebRequest -Uri "$baseUrl/v1/validate" -Method POST -Body $body -ContentType "application/json" -UseBasicParsing
    Write-Host "✅ Status: $($response.StatusCode)" -ForegroundColor Green
    Write-Host "📄 Resposta: $($response.Content)" -ForegroundColor Green
} catch {
    Write-Host "❌ Erro: $_" -ForegroundColor Red
}

Write-Host ""
Read-Host "Pressione Enter para fechar"
