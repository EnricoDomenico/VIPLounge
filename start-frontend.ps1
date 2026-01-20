# Script para iniciar o frontend
# Uso: .\start-frontend.ps1

Write-Host "╔═══════════════════════════════════════════╗" -ForegroundColor Green
Write-Host "║   🌐 Iniciando Frontend VIP Lounge        ║" -ForegroundColor Green
Write-Host "╚═══════════════════════════════════════════╝" -ForegroundColor Green
Write-Host ""

$port = 5000

# Verificar se Python está instalado
$pythonExists = Get-Command python -ErrorAction SilentlyContinue

if ($pythonExists) {
    Write-Host "✅ Python encontrado!" -ForegroundColor Green
    Write-Host "📌 Porta: $port" -ForegroundColor Cyan
    Write-Host "📂 Servindo: $PWD\web" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
    Write-Host ""
    Write-Host "🌐 Abra no navegador: http://localhost:$port" -ForegroundColor Yellow
    Write-Host "🛑 Pressione Ctrl+C para parar" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
    Write-Host ""
    
    Set-Location web
    python -m http.server $port
} else {
    Write-Host "❌ Python não encontrado!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Opções alternativas:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "1. Instalar Python:" -ForegroundColor Cyan
    Write-Host "   https://www.python.org/downloads/" -ForegroundColor White
    Write-Host ""
    Write-Host "2. Usar VS Code Live Server:" -ForegroundColor Cyan
    Write-Host "   - Abrir web/index.html no VS Code" -ForegroundColor White
    Write-Host "   - Clicar direito > Open with Live Server" -ForegroundColor White
    Write-Host ""
    Write-Host "3. Usar Node.js http-server:" -ForegroundColor Cyan
    Write-Host "   npm install -g http-server" -ForegroundColor White
    Write-Host "   cd web" -ForegroundColor White
    Write-Host "   http-server -p 5000" -ForegroundColor White
    Write-Host ""
}

Read-Host "Pressione Enter para fechar"
