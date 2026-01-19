# ✅ CHECKLIST DE SEGURANÇA - PRÉ-PRODUÇÃO

**VIP Lounge Platform**  
**Data:** 19 de Janeiro de 2026

---

## 🚨 CRÍTICO - BLOQUEIA DEPLOY

### Regenerar Tokens
- [ ] Ir em https://central.superlogica.net
- [ ] Gerar novo SUPERLOGICA_APP_TOKEN
- [ ] Gerar novo SUPERLOGICA_ACCESS_TOKEN
- [ ] Ir em https://app.clubeparcerias.com.br
- [ ] Regenerar REDE_PARCERIAS_BEARER_TOKEN
- [ ] Guardar tokens em local seguro (não GitHub!)

### Secrets no Google Cloud
- [ ] `gcloud auth login` (autenticar)
- [ ] Criar app-superlogica-app-token
- [ ] Criar app-superlogica-access-token
- [ ] Criar app-rede-parcerias-bearer-token
- [ ] Verificar secrets foram criados: `gcloud secrets list`

### Git Seguro
- [ ] Verificar `.env` NÃO será commitado
- [ ] `.gitignore` contém `.env` ✓
- [ ] Nenhum token em `git status`
- [ ] Nenhum token em `git diff --cached`
- [ ] Executar `bash scripts/security-check.sh` ✓

### Firestore Rules
- [ ] Acessar https://console.firebase.google.com
- [ ] Ir em Firestore Database > Rules
- [ ] Copiar conteúdo de `firestore.rules`
- [ ] Publicar rules

---

## 🔒 SEGURANÇA - IMPLEMENTAÇÃO

### Headers HTTP
- [x] X-Frame-Options (DENY)
- [x] X-Content-Type-Options (nosniff)
- [x] X-XSS-Protection (1; mode=block)
- [x] Strict-Transport-Security (HSTS)
- [x] Content-Security-Policy
- [x] Referrer-Policy
- [x] Permissions-Policy

### CORS
- [ ] Configurar domínio específico em `.env` (produção)
- [ ] NÃO usar wildcard `*` em produção
- [ ] Exemplo: `CORS_ORIGINS=https://seu-dominio.com`

### Validação
- [x] CPF validado com regex
- [x] Entrada sanitizada
- [x] Sem SQL injection
- [x] Sem command injection

### Frontend
- [x] Sem console.log() com dados sensíveis
- [x] Sem tokens no JavaScript
- [x] CPF mascarado
- [x] Sem XSS vulnerabilities

### Backend
- [x] Nenhum token hardcoded
- [x] Credenciais via env vars
- [x] Logging sem dados sensíveis
- [x] Error handling seguro

### Monitoramento
- [ ] Ativar Cloud Logging
- [ ] Configurar alertas
- [ ] Setup Cloud Monitoring
- [ ] Backup automático do Firestore

---

## 📋 DOCUMENTAÇÃO

- [x] SECURITY_SUMMARY.md criado
- [x] SECURITY_AUDIT.md criado
- [x] SECURITY_ANALYSIS_REPORT.md criado
- [x] GIT_COMMIT_GUIDE.md criado
- [x] scripts/security-check.sh criado
- [x] scripts/production-setup.sh criado
- [x] firestore.rules criado
- [x] internal/middleware/security.go criado

---

## 🚀 DEPLOYMENT

### Pré-Deploy
- [ ] Regenerar todos os tokens (visto acima)
- [ ] Criar secrets no Google Cloud (visto acima)
- [ ] Executar `bash scripts/security-check.sh`
- [ ] Revisar relatório de segurança
- [ ] Testar em staging

### Deploy
- [ ] Executar `bash scripts/production-setup.sh`
- [ ] Verificar `gcloud run services describe viplounge-prod`
- [ ] Testar acesso em https://seu-dominio.com
- [ ] Verificar HTTPS está ativado
- [ ] Confirmar CORS restrito

### Pós-Deploy
- [ ] Monitorar logs por 24h
- [ ] Verificar alertas funcionando
- [ ] Testar backup do Firestore
- [ ] Validar rate limiting

---

## 🔍 VERIFICAÇÕES FINAIS

### Tokens
- [ ] Nenhum token em `.git/` (executar: `git log --all -p | grep token`)
- [ ] `.env` não será commitado (executar: `git ls-files | grep .env`)
- [ ] Tokens em Secret Manager (executar: `gcloud secrets list`)

### Código
- [ ] Nenhum console.log() com CPF/dados
- [ ] Nenhum token hardcoded
- [ ] Nenhuma credencial em arquivos

### Infraestrutura
- [ ] Cloud Run com secrets configurados
- [ ] Firestore com rules restritivas
- [ ] CORS restrito a domínio
- [ ] HTTPS ativado
- [ ] Audit logs ativados

### Documentação
- [ ] README.md atualizado
- [ ] Guias de segurança lidos
- [ ] Checklist completo

---

## 📊 RESULTADOS

### Antes da Correção
```
❌ Credenciais: Expostas no .env
❌ CORS: Wildcard "*"
❌ HTTPS: Não obrigatório  
❌ Firestore: Sem proteção
❌ Status: CRÍTICO - NÃO FAZER DEPLOY
```

### Depois da Correção
```
✅ Credenciais: Secret Manager
✅ CORS: Domínio específico
✅ HTTPS: Cloud Run (automático)
✅ Firestore: Rules restritivas
✅ Status: PRONTO PARA PRODUÇÃO
```

---

## 🎯 TIMELINE

```
DIA 1 (Hoje):
├─ ☐ Regenerar tokens
├─ ☐ Criar secrets
├─ ☐ Ler relatórios
└─ ☐ Fazer verificações

DIA 2:
├─ ☐ Configurar Firestore Rules
├─ ☐ Testar em staging
├─ ☐ Revisar logs
└─ ☐ Fazer deploy

DIA 3+:
├─ ☐ Monitorar produção
├─ ☐ Verificar alertas
├─ ☐ Validar backups
└─ ☐ Documentar
```

---

## 💡 REMINDERS

```
🔐 Uma credencial vazada = CPF de usuários em risco!

❌ NUNCA commitar .env com tokens reais
❌ NUNCA usar CORS wildcard em produção
❌ NUNCA deixar Firestore sem rules

✅ SEMPRE regenerar tokens
✅ SEMPRE usar Secret Manager
✅ SEMPRE testar em staging
```

---

## 📞 PRÓXIMO PASSO

**Leia nesta ordem:**

1. [SECURITY_SUMMARY.md](SECURITY_SUMMARY.md) - Resumo (2 min)
2. [GIT_COMMIT_GUIDE.md](GIT_COMMIT_GUIDE.md) - Como fazer commit (5 min)
3. [SECURITY_ANALYSIS_REPORT.md](SECURITY_ANALYSIS_REPORT.md) - Detalhes (15 min)
4. [SECURITY_AUDIT.md](SECURITY_AUDIT.md) - Técnico (20 min)

---

**Status: 🔴 AGUARDANDO IMPLEMENTAÇÃO**

Após completar todos os itens: ✅ **PRONTO PARA PRODUÇÃO**
