# ⚡ QUICK TEST - Validação do Backend

## 1️⃣ Verificar CORS Headers

```bash
# Testar requisição com headers CORS
curl -i -X OPTIONS http://localhost:8080/api/v1/ \
  -H "Origin: http://localhost:5000" \
  -H "Access-Control-Request-Method: GET"
```

**Esperado:** Header `Access-Control-Allow-Origin: http://localhost:5000`

---

## 2️⃣ Testar Requisição de API Local

```bash
# Testar GET
curl -X GET http://localhost:8080/api/v1/health

# Testar POST
curl -X POST http://localhost:8080/api/v1/validation \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}'
```

---

## 3️⃣ Verificar Firebase.json

Visualizar a configuração de rewrites:

```bash
cat firebase.json | grep -A 10 "rewrites"
```

**Esperado:**
```json
"rewrites": [
  {"source": "/api/**", "function": "viplounge-service"},
  {"source": "**", "destination": "/index.html"}
]
```

---

## 4️⃣ Testar no Navegador

1. Abrir Console do Navegador (F12)
2. Ir para a aba **Network**
3. Clicar em qualquer botão que faça requisição de API
4. Verificar:
   - ✅ Status 200/201 (não 404)
   - ✅ Response não é HTML da SPA
   - ✅ Headers têm `Access-Control-Allow-Origin`

---

## 5️⃣ Verificar Logs do Backend

```bash
# Ver últimas linhas dos logs
go run cmd/server/main.go

# Ou no Cloud Run
gcloud run logs read viplounge-service --limit=50
```

**Esperado:**
```
🚀 Server 'VIP Lounge' starting on port 8080
[Requisições entrando...]
```

---

## 6️⃣ Testar Autenticação Firebase

```bash
# Verificar se Firebase está conectado
curl -X GET http://localhost:8080/api/v1/user \
  -H "Authorization: Bearer seu_token_firebase"
```

---

## 🆘 Se Ainda não Funcionar

### Erro 404 em `/api`
- Verificar [internal/handler/http.go](internal/handler/http.go) - rotas configuradas?

### Erro CORS
- Verificar origem do navegador
- Adicionar em [cmd/server/main.go](cmd/server/main.go) `corsMiddleware`

### Erro Firestore
- Verificar `GOOGLE_CLOUD_PROJECT` env var
- Testar: `echo $GOOGLE_CLOUD_PROJECT`

### Erro "Connection Refused"
- Backend não está rodando?
- `go run cmd/server/main.go`

---

**Comando tudo-em-um para testar:**

```bash
# Supondo que backend está rodando em http://localhost:8080
curl -v http://localhost:8080/api/v1/health -H "Origin: http://localhost:5000"
```
