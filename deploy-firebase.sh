#!/bin/bash

# Script para fazer deploy no Firebase (Functions + Hosting)
# Uso: ./deploy-firebase.sh

set -e

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         🚀 FIREBASE DEPLOY (Functions + Hosting)           ║"
echo "╚════════════════════════════════════════════════════════════╝"

# Verificar se Firebase CLI está instalado
if ! command -v firebase &> /dev/null; then
    echo "❌ Firebase CLI não encontrado. Instale com:"
    echo "   npm install -g firebase-tools"
    exit 1
fi

# Fazer deploy
echo "📦 Deployando Functions + Hosting..."
firebase deploy --only functions,hosting

echo ""
echo "✅ Deploy concluído com sucesso!"
echo ""
echo "🔗 URLs:"
echo "   Firebase Hosting: https://viplounge-f079a.firebaseapp.com"
echo "   Cloud Function:   https://us-central1-viplounge-f079a.cloudfunctions.net/apiProxy"
echo ""
echo "📝 Logs:"
echo "   firebase functions:log"
echo "   firebase hosting:logs"
