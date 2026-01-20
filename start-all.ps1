# Script COMPLETO para iniciar TODO o projeto
# Uso: .\start-all.ps1

Write-Host "╔═══════════════════════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║   🚀 Iniciando VIP Lounge - Backend + Frontend           ║" -ForegroundColor Green
Write-Host "╚═══════════════════════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

# Função para iniciar processo em nova janela
function Start-InNewWindow {
    param(
        [string]$Title,
        [string]$Command
    )
    
    $encodedCommand = [Convert]::ToBase64String([System.Text.Encoding]::Unicode.GetBytes($Command))
    Start-Process powershell -ArgumentList "-NoExit", "-EncodedCommand", $encodedCommand -WindowStyle Normal
    Write-Host "✅ $Title iniciado em nova janela" -ForegroundColor Green
}

Write-Host "📋 Iniciando serviços..." -ForegroundColor Cyan
Write-Host ""

# 1. Iniciar Backend
Write-Host "1️⃣  Iniciando Backend (porta 8081)..." -ForegroundColor Yellow
$backendCommand = @"
`$host.UI.RawUI.WindowTitle = 'Backend - VIP Lounge'
Set-Location 'b:\Games\viplounge'
`$env:PORT = '8081'
Write-Host '🚀 Iniciando Backend...' -ForegroundColor Green
go run cmd/server/main.go
"@
Start-InNewWindow "Backend" $backendCommand
Start-Sleep -Seconds 2

# 2. Iniciar Frontend
Write-Host "2️⃣  Iniciando Frontend (porta 5000)..." -ForegroundColor Yellow

# Verificar se Python existe
$pythonExists = Get-Command python -ErrorAction SilentlyContinue

if ($pythonExists) {
    $frontendCommand = @"
`$host.UI.RawUI.WindowTitle = 'Frontend - VIP Lounge'
Set-Location 'b:\Games\viplounge\web'
Write-Host '🌐 Iniciando Frontend...' -ForegroundColor Green
Write-Host ''
Write-Host '🌐 Acesse: http://localhost:5000' -ForegroundColor Yellow
Write-Host '🛑 Pressione Ctrl+C para parar' -ForegroundColor Yellow
Write-Host ''
python -m http.server 5000
"@
    Start-InNewWindow "Frontend" $frontendCommand
    Start-Sleep -Seconds 2
    
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
    Write-Host ""
    Write-Host "✅ Serviços iniciados com sucesso!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📊 Status:" -ForegroundColor Cyan
    Write-Host "   Backend:  http://localhost:8081" -ForegroundColor White
    Write-Host "   Frontend: http://localhost:5000" -ForegroundColor White
    Write-Host ""
    Write-Host "🌐 Abra no navegador: http://localhost:5000" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "❌ Python não encontrado!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Para continuar:" -ForegroundColor Yellow
    Write-Host "1. Abrir web/index.html no VS Code" -ForegroundColor White
    Write-Host "2. Clicar direito > Open with Live Server" -ForegroundColor White
    Write-Host ""
}

Write-Host "💡 Duas janelas PowerShell foram abertas (Backend e Frontend)" -ForegroundColor Cyan
Write-Host "   Não feche essas janelas!" -ForegroundColor Yellow
Write-Host ""

Read-Host "Pressione Enter para fechar esta janela (serviços continuarão rodando)"
