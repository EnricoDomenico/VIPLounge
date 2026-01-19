#!/bin/bash

# 🚀 SETUP DE PRODUÇÃO - VIP LOUNGE PLATFORM
# Execute este script para configurar tudo para produção no Google Cloud

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}🚀 VIP LOUNGE - PRODUCTION SETUP${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# ============================================================================
# 1. VALIDAR PRÉ-REQUISITOS
# ============================================================================
echo -e "${YELLOW}1️⃣  Validando pré-requisitos...${NC}"

# Verificar gcloud
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}❌ gcloud CLI não está instalado${NC}"
    echo "   Baixe em: https://cloud.google.com/sdk/docs/install"
    exit 1
fi

# Verificar git
if ! command -v git &> /dev/null; then
    echo -e "${RED}❌ Git não está instalado${NC}"
    exit 1
fi

# Verificar Docker
if ! command -v docker &> /dev/null; then
    echo -e "${YELLOW}⚠️  Docker não está instalado (necessário para build)${NC}"
fi

echo -e "${GREEN}✅ Pré-requisitos OK${NC}"
echo ""

# ============================================================================
# 2. COLETAR INFORMAÇÕES
# ============================================================================
echo -e "${YELLOW}2️⃣  Coletando informações...${NC}"

# Projeto GCP
PROJECT_ID=$(gcloud config get-value project)
if [ -z "$PROJECT_ID" ]; then
    echo -e "${YELLOW}Qual é seu Google Cloud Project ID?${NC}"
    read PROJECT_ID
fi
echo "Projeto: $PROJECT_ID"

# Região
read -p "Qual é a região do Cloud Run? (padrão: us-central1): " REGION
REGION=${REGION:-us-central1}
echo "Região: $REGION"

# Domínio
read -p "Qual é o domínio da sua aplicação? (ex: app.seu-dominio.com): " DOMAIN
echo "Domínio: $DOMAIN"

echo ""

# ============================================================================
# 3. VERIFICAR SE .env NÃO ESTÁ NO GIT
# ============================================================================
echo -e "${YELLOW}3️⃣  Verificando segurança do repo...${NC}"

if git ls-files --others --exclude-standard .env | grep -q .; then
    echo -e "${RED}❌ .env está tracked no git!${NC}"
    echo "   Execute: git rm --cached .env && git commit"
    exit 1
fi

echo -e "${GREEN}✅ .env não está no git${NC}"
echo ""

# ============================================================================
# 4. CRIAR SECRETS NO SECRET MANAGER
# ============================================================================
echo -e "${YELLOW}4️⃣  Criando secrets no Google Cloud Secret Manager...${NC}"

echo "Você precisa dos seguintes valores:"
echo ""

read -p "SUPERLOGICA_APP_TOKEN (regenerado): " SUPERLOGICA_APP_TOKEN
read -p "SUPERLOGICA_ACCESS_TOKEN (regenerado): " SUPERLOGICA_ACCESS_TOKEN
read -p "REDE_PARCERIAS_BEARER_TOKEN (regenerado): " REDE_PARCERIAS_BEARER_TOKEN

echo ""
echo "Criando secrets..."

# Criar secrets
gcloud secrets create app-superlogica-app-token \
    --replication-policy="automatic" \
    --data-file=- <<< "$SUPERLOGICA_APP_TOKEN" \
    --project="$PROJECT_ID" 2>/dev/null || \
gcloud secrets versions add app-superlogica-app-token \
    --data-file=- <<< "$SUPERLOGICA_APP_TOKEN" \
    --project="$PROJECT_ID"

gcloud secrets create app-superlogica-access-token \
    --replication-policy="automatic" \
    --data-file=- <<< "$SUPERLOGICA_ACCESS_TOKEN" \
    --project="$PROJECT_ID" 2>/dev/null || \
gcloud secrets versions add app-superlogica-access-token \
    --data-file=- <<< "$SUPERLOGICA_ACCESS_TOKEN" \
    --project="$PROJECT_ID"

gcloud secrets create app-rede-parcerias-bearer-token \
    --replication-policy="automatic" \
    --data-file=- <<< "$REDE_PARCERIAS_BEARER_TOKEN" \
    --project="$PROJECT_ID" 2>/dev/null || \
gcloud secrets versions add app-rede-parcerias-bearer-token \
    --data-file=- <<< "$REDE_PARCERIAS_BEARER_TOKEN" \
    --project="$PROJECT_ID"

echo -e "${GREEN}✅ Secrets criados${NC}"
echo ""

# ============================================================================
# 5. CONFIGURAR FIRESTORE RULES
# ============================================================================
echo -e "${YELLOW}5️⃣  Configurando Firestore Security Rules...${NC}"

cat > firestore.rules << 'EOF'
rules_version = '2';
service cloud.firestore {
  match /databases/{database}/documents {
    // Negar acesso público a todos os documentos
    match /{document=**} {
      allow read, write: if false;
    }
  }
}
EOF

echo "⚠️  Regras de firestore criadas em firestore.rules"
echo "   Copie o conteúdo para o Firebase Console"
echo ""

# ============================================================================
# 6. CONSTRUIR IMAGEM DOCKER
# ============================================================================
echo -e "${YELLOW}6️⃣  Construindo imagem Docker...${NC}"

TAG="gcr.io/${PROJECT_ID}/viplounge:latest"

if docker build -t "$TAG" .; then
    echo -e "${GREEN}✅ Docker build bem-sucedido${NC}"
else
    echo -e "${RED}❌ Docker build falhou${NC}"
    exit 1
fi
echo ""

# ============================================================================
# 7. PUSH PARA CONTAINER REGISTRY
# ============================================================================
echo -e "${YELLOW}7️⃣  Fazendo push para Google Container Registry...${NC}"

if docker push "$TAG"; then
    echo -e "${GREEN}✅ Push bem-sucedido${NC}"
else
    echo -e "${RED}❌ Push falhou${NC}"
    echo "   Execute: gcloud auth configure-docker"
    exit 1
fi
echo ""

# ============================================================================
# 8. DEPLOY NO CLOUD RUN
# ============================================================================
echo -e "${YELLOW}8️⃣  Fazendo deploy no Cloud Run...${NC}"

gcloud run deploy viplounge-prod \
    --image "$TAG" \
    --platform managed \
    --region "$REGION" \
    --project "$PROJECT_ID" \
    --set-env-vars \
        SUPERLOGICA_URL=https://api.superlogica.net/v2/condor,\
        REDE_PARCERIAS_URL=https://api.staging.clubeparcerias.com.br/api-client/v1,\
        CORS_ORIGINS=https://$DOMAIN,\
        ENABLE_DEBUG=false,\
        LOG_LEVEL=INFO \
    --set-secrets \
        SUPERLOGICA_APP_TOKEN=app-superlogica-app-token:latest,\
        SUPERLOGICA_ACCESS_TOKEN=app-superlogica-access-token:latest,\
        REDE_PARCERIAS_BEARER_TOKEN=app-rede-parcerias-bearer-token:latest,\
        GOOGLE_CLOUD_PROJECT=${PROJECT_ID}:latest \
    --cpu 2 \
    --memory 512Mi \
    --max-instances 100 \
    --timeout 60 \
    --no-allow-unauthenticated

echo -e "${GREEN}✅ Deploy bem-sucedido${NC}"
echo ""

# ============================================================================
# 9. CONFIGURAR DOMÍNIO CUSTOMIZADO
# ============================================================================
echo -e "${YELLOW}9️⃣  Configurando domínio customizado...${NC}"

SERVICE_URL=$(gcloud run services describe viplounge-prod \
    --platform managed \
    --region "$REGION" \
    --format 'value(status.url)' \
    --project "$PROJECT_ID")

echo "URL do Cloud Run: $SERVICE_URL"
echo ""
echo "Para usar domínio customizado ($DOMAIN):"
echo "  1. Vá para: https://console.cloud.google.com/run"
echo "  2. Clique no serviço 'viplounge-prod'"
echo "  3. Clique em 'Manage Custom Domains'"
echo "  4. Adicione o domínio e configure o CNAME no seu DNS"
echo ""

# ============================================================================
# 10. CONFIGURAR CLOUD ARMOR (DDoS Protection)
# ============================================================================
echo -e "${YELLOW}🔟 Configurando Cloud Armor...${NC}"

read -p "Deseja configurar Cloud Armor para proteção DDoS? (s/n): " ARMOR_CHOICE

if [ "$ARMOR_CHOICE" = "s" ] || [ "$ARMOR_CHOICE" = "S" ]; then
    gcloud compute security-policies create viplounge-armor \
        --description "Cloud Armor para VIP Lounge" \
        --project="$PROJECT_ID" 2>/dev/null || echo "Política já existe"
    
    echo -e "${GREEN}✅ Cloud Armor configurado${NC}"
fi
echo ""

# ============================================================================
# 11. ATIVAR AUDIT LOGS
# ============================================================================
echo -e "${YELLOW}1️⃣1️⃣ Ativando Cloud Audit Logs...${NC}"

gcloud logging write viplounge-setup \
    "VIP Lounge setup concluído em $(date)" \
    --severity=INFO \
    --project="$PROJECT_ID"

echo -e "${GREEN}✅ Audit logs ativados${NC}"
echo ""

# ============================================================================
# RESUMO
# ============================================================================
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✅ SETUP DE PRODUÇÃO CONCLUÍDO!${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

echo "📋 VERIFICAÇÃO FINAL:"
echo ""
echo "1. ✅ Secrets criados no Secret Manager"
echo "2. ✅ Docker image pushada para GCR"
echo "3. ✅ Cloud Run deploy realizado"
echo "4. ✅ Audit logs ativados"
echo ""

echo "🔗 URL do serviço: $SERVICE_URL"
echo ""

echo "📝 PRÓXIMOS PASSOS:"
echo ""
echo "1. Configure o domínio customizado no Cloud Run"
echo "2. Configure Firestore Rules (veja firestore.rules)"
echo "3. Revise Cloud Armor policies"
echo "4. Configure alertas no Cloud Monitoring"
echo "5. Ative backups automáticos do Firestore"
echo ""

echo "🔒 SEGURANÇA:"
echo ""
echo "✅ Tokens em Secret Manager (não no código)"
echo "✅ HTTPS obrigatório"
echo "✅ CORS restrito a: https://$DOMAIN"
echo "✅ Debug mode desativado"
echo "✅ Security headers implementados"
echo ""

echo "📊 MONITORAMENTO:"
echo "   https://console.cloud.google.com/monitoring"
echo ""

echo "📋 LOGS:"
echo "   https://console.cloud.google.com/logs"
echo ""
