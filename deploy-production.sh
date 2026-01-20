#!/bin/bash

# VIP Lounge - Complete Firebase + Backend Deployment Script
# Conecta Frontend (Firebase Hosting) com Backend (Cloud Run) e Firestore

set -e

echo "🚀 VIP LOUNGE - DEPLOYMENT SETUP"
echo "=================================="
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configurações
PROJECT_ID="viplounge-f079a"
REGION="southamerica-east1"
SERVICE_NAME="viplounge-backend"
FRONTEND_URL="https://viplounge-f079a.web.app"

echo -e "${BLUE}1. Verificando configuração do Firebase...${NC}"
firebase projects:list | grep $PROJECT_ID
echo -e "${GREEN}✓ Projeto Firebase OK${NC}\n"

echo -e "${BLUE}2. Verificando Google Cloud SDK...${NC}"
if ! command -v gcloud &> /dev/null; then
    echo -e "${RED}✗ gcloud CLI não encontrado!${NC}"
    echo "   Instale: https://cloud.google.com/sdk/docs/install"
    exit 1
fi
echo -e "${GREEN}✓ gcloud CLI disponível${NC}\n"

echo -e "${BLUE}3. Autenticando no Google Cloud...${NC}"
gcloud auth login
gcloud config set project $PROJECT_ID
echo -e "${GREEN}✓ Autenticação OK${NC}\n"

echo -e "${BLUE}4. Criando Secrets no Google Cloud Secret Manager...${NC}"
# Lê credenciais do .env (assumindo que já foram preparadas)
if [ -f ".env" ]; then
    echo "   Carregando credenciais de .env..."
    export $(cat .env | xargs)
    
    # Criar secrets
    echo "$SUPERLOGICA_APP_TOKEN" | gcloud secrets create superlogica-app-token --data-file=- 2>/dev/null || echo "   Secret superlogica-app-token já existe"
    echo "$SUPERLOGICA_ACCESS_TOKEN" | gcloud secrets create superlogica-access-token --data-file=- 2>/dev/null || echo "   Secret superlogica-access-token já existe"
    echo "$REDE_PARCERIAS_BEARER_TOKEN" | gcloud secrets create rede-parcerias-bearer-token --data-file=- 2>/dev/null || echo "   Secret rede-parcerias-bearer-token já existe"
    
    echo -e "${GREEN}✓ Secrets criados no Secret Manager${NC}\n"
else
    echo -e "${RED}✗ Arquivo .env não encontrado!${NC}"
    echo "   Crie um .env com as credenciais (não fazer commit!)"
    exit 1
fi

echo -e "${BLUE}5. Deploying Backend no Cloud Run...${NC}"
gcloud run deploy $SERVICE_NAME \
    --source . \
    --region $REGION \
    --allow-unauthenticated \
    --platform managed \
    --set-env-vars="CORS_ORIGINS=$FRONTEND_URL,REQUIRE_HTTPS=true,ENABLE_DEBUG=false" \
    --update-secrets="SUPERLOGICA_APP_TOKEN=superlogica-app-token:latest,SUPERLOGICA_ACCESS_TOKEN=superlogica-access-token:latest,REDE_PARCERIAS_BEARER_TOKEN=rede-parcerias-bearer-token:latest"

BACKEND_URL=$(gcloud run services describe $SERVICE_NAME --region $REGION --format 'value(status.url)')
echo -e "${GREEN}✓ Backend deployed: $BACKEND_URL${NC}\n"

echo -e "${BLUE}6. Configurando frontend para se conectar ao backend...${NC}"
# Criar arquivo de config com URL do backend
cat > web/config.js << EOF
const CONFIG = {
  API_BASE_URL: '$BACKEND_URL',
  FRONTEND_URL: '$FRONTEND_URL',
  PROJECT_ID: '$PROJECT_ID'
};
EOF
echo -e "${GREEN}✓ Arquivo web/config.js criado${NC}\n"

echo -e "${BLUE}7. Deploying Frontend (Firebase Hosting)...${NC}"
firebase deploy --only hosting

echo -e "${GREEN}✓ Frontend deployed${NC}\n"

echo -e "${BLUE}8. Configurando Firestore Security Rules...${NC}"
firebase deploy --only firestore:rules

echo -e "${GREEN}✓ Firestore Rules deployed${NC}\n"

echo ""
echo "=========================================="
echo -e "${GREEN}✅ DEPLOYMENT COMPLETO!${NC}"
echo "=========================================="
echo ""
echo "📍 URLs:"
echo -e "   Frontend:  ${GREEN}$FRONTEND_URL${NC}"
echo -e "   Backend:   ${GREEN}$BACKEND_URL${NC}"
echo -e "   Firestore: ${GREEN}https://console.firebase.google.com/project/$PROJECT_ID/firestore${NC}"
echo -e "   Cloud Run: ${GREEN}https://console.cloud.google.com/run?project=$PROJECT_ID${NC}"
echo ""
echo "🔒 Segurança:"
echo "   ✓ CORS configurado para frontend apenas"
echo "   ✓ HTTPS obrigatório"
echo "   ✓ Credenciais em Secret Manager"
echo "   ✓ Firestore com rules restritivas"
echo ""
echo "🧪 Testes recomendados:"
echo "   1. Abra: $FRONTEND_URL"
echo "   2. Teste CPF validation"
echo "   3. Verifique console do navegador"
echo "   4. Monitore: Cloud Run dashboard"
echo ""
echo "📊 Próximos passos:"
echo "   - Revisar logs em Cloud Run"
echo "   - Configurar alertas em Cloud Monitoring"
echo "   - Backup automático do Firestore"
echo "   - Testes de carga/stress"
echo ""
