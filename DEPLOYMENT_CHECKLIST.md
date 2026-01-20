# 🚀 DEPLOYMENT CHECKLIST

## Status: ✅ PRONTO PARA DEPLOY

As seguintes correções foram implementadas:

### ✅ Modificações Realizadas

1. **firebase.json** - Adicionado rewrite para `/api/**` → Cloud Run
2. **cmd/server/main.go** - Adicionado middleware CORS
3. **web/api-config.js** - Atualizado para usar same-origin
4. **cloudbuild.yaml** - Adicionado GOOGLE_CLOUD_PROJECT env var

---

## 📝 PRÉ-REQUISITOS

- [ ] Git commit das mudanças
- [ ] Projeto GCP configurado
- [ ] Cloud Run habilitado
- [ ] Firestore Database criado
- [ ] Firebase Hosting conectado

---

## 🔄 PASSO A PASSO DO DEPLOY

### 1. Commit e Push
```bash
git add .
git commit -m "fix: corrigir CORS e roteamento de API no Firebase"
git push origin main
```

### 2. Trigger Cloud Build
```bash
# Automático via webhook, ou manual:
gcloud builds submit --config cloudbuild.yaml
```

### 3. Monitorar Build
```bash
# Ver status do build
gcloud builds log $(gcloud builds list --limit=1 --format='value(ID)')

# Ou web console
# https://console.cloud.google.com/cloud-build
```

### 4. Verificar Deploy no Cloud Run
```bash
gcloud run services describe viplounge-service --region us-central1
```

Procurar por:
- Status: `ACTIVE`
- URL: `https://viplounge-service-xxxxx.run.app`

### 5. Teste de Conectividade
```bash
# Pegar URL do Cloud Run
SERVICE_URL=$(gcloud run services describe viplounge-service \
  --region us-central1 --format='value(status.url)')

# Testar
curl -X GET $SERVICE_URL/api/v1/health
```

### 6. Deploy do Frontend no Firebase
```bash
firebase deploy --only hosting
```

---

## ✨ VERIFICAÇÕES PÓS-DEPLOY

### No Cloud Console
- [ ] Cloud Run: status `ACTIVE`
- [ ] Cloud Build: build bem-sucedido
- [ ] Firebase Hosting: hospedagem ativa
- [ ] Firestore: database conectado

### No Navegador
1. Abrir: https://seu-projeto.firebaseapp.com
2. Abrir DevTools (F12)
3. Aba **Network**
4. Fazer ação que chama API
5. Verificar:
   - [ ] Requisição `/api/v1/*` retorna 200/201
   - [ ] Response é JSON (não HTML)
   - [ ] Sem erros CORS

### Via Terminal
```bash
# Testar endpoint do Cloud Run
curl https://seu-projeto.firebaseapp.com/api/v1/health

# Ou diretamente
curl https://viplounge-service-xxxxx.run.app/api/v1/health
```

---

## 🆘 TROUBLESHOOTING

### Cloud Build Falha
```bash
# Ver logs detalhados
gcloud builds log <BUILD_ID> --stream
```

### Cloud Run retorna 500
```bash
# Ver logs do serviço
gcloud run logs read viplounge-service --limit=100
```

### CORS ainda falhando
- Verificar origin do site
- Adicionar em `allowedOrigins` map
- Redeploy

### Firestore não conecta
- Verificar `GOOGLE_CLOUD_PROJECT` set no Cloud Run
- Verificar permissões do serviço account

---

## 📊 MONITORAMENTO

Após deploy, monitorar:

```bash
# Logs em tempo real
gcloud run logs read viplounge-service --tail

# Métricas
gcloud monitoring dashboards list

# Erros
gcloud error-reporting list
```

---

## 🎯 RESUMO DO FLUXO

```
Frontend (Firebase Hosting)
    ↓ GET /api/v1/validation
Firebase.json rewrite /api/** → Cloud Run
    ↓
Cloud Run (viplounge-service)
    ↓ corsMiddleware valida origin
    ↓ Handler processa
    ↓ Firestore query
    ↓ Response JSON 200
    ↓
Frontend recebe dados
```

---

**Última atualização:** 20/01/2026  
**Status:** ✅ READY FOR DEPLOYMENT
