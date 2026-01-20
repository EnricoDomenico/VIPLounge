#!/bin/bash

# Script para descobrir URL do Cloud Run
# Requer: gcloud CLI instalado e autenticado

echo "🔍 Descobrindo URL do Cloud Run..."
echo ""

# Lista todos os serviços Cloud Run
echo "📋 Serviços Cloud Run disponíveis:"
gcloud run services list --platform managed --format="table(SERVICE,REGION,URL)"

echo ""
echo "📌 URL do viplounge-service:"
gcloud run services describe viplounge-service --region us-central1 --format="value(status.url)"

echo ""
echo "✅ Copie a URL acima e cole em web/backend-config.json"
echo "   Exemplo: {\"backendUrl\": \"https://viplounge-service-xxx.run.app\"}"
