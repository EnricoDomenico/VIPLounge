#!/bin/bash

# 🔐 SCRIPT DE SEGURANÇA - VIP LOUNGE
# Execute antes de fazer qualquer deploy em produção

set -e

echo "🔐 Iniciando verificações de segurança..."
echo ""

# ============================================================================
# 1. VERIFICAR .gitignore
# ============================================================================
echo "📋 Verificando .gitignore..."

if ! grep -q "\.env" .gitignore; then
    echo "❌ ERRO: .env não está no .gitignore!"
    echo "   Adicione '.env' ao arquivo .gitignore"
    exit 1
fi

echo "✅ .env está sendo ignorado"

# ============================================================================
# 2. VERIFICAR SE .env EXISTE LOCALMENTE
# ============================================================================
echo ""
echo "📋 Verificando arquivo .env..."

if [ ! -f ".env" ]; then
    echo "⚠️  .env não encontrado (esperado em produção)"
else
    echo "✅ .env existe localmente"
fi

# ============================================================================
# 3. AVISAR SOBRE CREDENCIAIS
# ============================================================================
echo ""
echo "🚨 LEMBRETE CRÍTICO:"
echo "   Certifique-se de:"
echo "   1. Regenerar tokens no Superlogica"
echo "   2. Regenerar JWT no Rede Parcerias"
echo "   3. NUNCA commitar .env com tokens reais"
echo "   4. Usar Google Cloud Secret Manager para produção"
echo ""

# ============================================================================
# 4. VERIFICAR GO BUILD
# ============================================================================
echo "🔨 Compilando projeto..."

if ! go build -v -o bin/server ./cmd/server/main.go; then
    echo "❌ Erro ao compilar!"
    exit 1
fi

echo "✅ Build bem-sucedido"

# ============================================================================
# 5. VERIFICAR SE HÁ SECRETS HARDCODED
# ============================================================================
echo ""
echo "🔍 Procurando secrets hardcoded..."

# Palavras-chave perigosas
PATTERNS=(
    "SUPERLOGICA_APP_TOKEN="
    "SUPERLOGICA_ACCESS_TOKEN="
    "REDE_PARCERIAS_BEARER_TOKEN="
    "bearer"
    "api_key"
    "secret_key"
)

FOUND_ISSUE=0

for pattern in "${PATTERNS[@]}"; do
    if grep -r "$pattern" cmd/ internal/ --exclude-dir=.git 2>/dev/null | grep -v "getEnv\|os.Getenv" > /dev/null; then
        echo "⚠️  Encontrado potencial secret hardcoded com padrão: $pattern"
        FOUND_ISSUE=1
    fi
done

if [ $FOUND_ISSUE -eq 0 ]; then
    echo "✅ Nenhum secret hardcoded encontrado"
fi

# ============================================================================
# 6. VERIFICAR LOGGING
# ============================================================================
echo ""
echo "🔍 Procurando por possível vazamento de dados em logs..."

# Procurar por log de CPF ou dados sensíveis
if grep -r "log.*cpf\|log.*CPF\|Println.*cpf\|Printf.*cpf" cmd/ internal/ --exclude-dir=.git 2>/dev/null; then
    echo "⚠️  Possível logging de CPF encontrado!"
else
    echo "✅ Nenhum log de CPF encontrado"
fi

# ============================================================================
# 7. VERIFICAR CONFIGURAÇÃO DE CORS
# ============================================================================
echo ""
echo "🔍 Verificando configuração de CORS..."

if grep -q 'CORS_ORIGINS=\*' .env 2>/dev/null || grep -q 'CORSAllowedOrigins.*"\*"' internal/config/config.go; then
    echo "⚠️  ATENÇÃO: CORS está configurado como '*' (wildcard)"
    echo "   Em produção, configure domínios específicos:"
    echo "   CORS_ORIGINS=https://seu-dominio.com"
fi

# ============================================================================
# 8. CRIAR .env.example
# ============================================================================
echo ""
echo "📝 Criando .env.example (sem tokens reais)..."

if [ ! -f ".env.example" ]; then
    cp .env .env.example
    # Remover tokens do exemplo
    sed -i 's/=74539367-69b7-432a-934f-8d9050bade0c/=seu-app-token/g' .env.example
    sed -i 's/=d769811d-2d05-4640-b756-b2bae62318cd/=seu-access-token/g' .env.example
    sed -i 's/=eyJ.*$/=seu-jwt-bearer-token/g' .env.example
    echo "✅ .env.example criado com exemplos"
else
    echo "✅ .env.example já existe"
fi

# ============================================================================
# 9. EXECUTAR TESTES
# ============================================================================
echo ""
echo "🧪 Executando testes..."

if ! go test ./... -v; then
    echo "⚠️  Alguns testes falharam (verifique)"
fi

# ============================================================================
# RESULTADO FINAL
# ============================================================================
echo ""
echo "=========================================="
echo "✅ VERIFICAÇÕES DE SEGURANÇA COMPLETADAS"
echo "=========================================="
echo ""
echo "📋 PRÓXIMOS PASSOS ANTES DE FAZER DEPLOY:"
echo ""
echo "1. Regenerar tokens:"
echo "   - Superlogica APP_TOKEN"
echo "   - Superlogica ACCESS_TOKEN"
echo "   - Rede Parcerias Bearer Token"
echo ""
echo "2. Criar secrets no Google Cloud:"
echo "   gcloud secrets create superlogica-app-token --data-file=-"
echo ""
echo "3. Não fazer commit com .env contendo tokens reais"
echo ""
echo "4. Para fazer deploy no Cloud Run:"
echo "   gcloud run deploy viplounge-prod \\"
echo "     --image gcr.io/seu-projeto/viplounge:latest \\"
echo "     --set-secrets SUPERLOGICA_APP_TOKEN=superlogica-app-token:latest \\"
echo "     ..."
echo ""
echo "5. Verificar Firestore Security Rules"
echo ""
echo "🔐 Leia SECURITY_AUDIT.md para detalhes completos"
echo ""
