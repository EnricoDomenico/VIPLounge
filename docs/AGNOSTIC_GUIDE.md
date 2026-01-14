# Guia de Configuração Agnóstica - VIP Lounge

## 🎯 O que é "Agnóstico"?

Agnóstico significa que a plataforma **não assume nada específico** sobre o cliente. Você pode customizar completamente:

- **Branding**: Nome da app, cores, logo
- **Mensagens**: Todos os textos em qualquer idioma
- **Comportamento**: Fluxo de UX, validações, redirecionamentos
- **Integrações**: APIs de validação e parceiros
- **Database**: Tipo de persistência

## 📋 Como Funciona?

### 1. **Configuração por Arquivo YAML** (`config.yaml`)

O arquivo `config.yaml` é carregado na inicialização e pode ser customizado sem tocar no código:

```yaml
branding:
  app_name: "Seu App"
  company_name: "Sua Empresa"
  theme_color: "FF5733"

messages:
  welcome_title: "Bem-vindo"
  success_message: "Sucesso!"

behavior:
  enable_debug_panel: false
  language: "en-US"
```

### 2. **Configuração por Variáveis de Ambiente** (`.env`)

As variáveis de ambiente **sobrescrevem** o `config.yaml`:

```bash
APP_NAME="Meu App Customizado"
COMPANY_EMAIL="contato@meuapp.com"
MSG_SUCCESS_MSG="Parabéns, você foi aprovado!"
```

### 3. **Frontend Dinâmico**

O frontend **carrega a config do servidor** via endpoint `/config`:

```javascript
// Chamada automática ao carregar
fetch('/config')
  .then(r => r.json())
  .then(cfg => {
    // Aplicar todas as customizações dinamicamente
    document.title = cfg.branding.app_name;
    // ... etc
  });
```

## 🔧 Exemplos de Customização

### Exemplo 1: Mudar Branding para Outro Cliente

**config.yaml:**
```yaml
branding:
  app_name: "Clube Prime"
  company_name: "Prime Benefícios"
  company_email: "suporte@prime.com"
  theme_color: "FF6B00"      # Laranja
  secondary_color: "FFB800"  # Ouro

messages:
  welcome_title: "Acesso Exclusivo"
  success_title: "APROVADO!"
  success_message: "Bem-vindo ao Clube Prime!"
```

**Resultado:** A app muda completamente de identidade visual e mensagens, sem modificar código Go.

---

### Exemplo 2: Mudar Idioma para Inglês

**`.env`:**
```bash
LANGUAGE=en-US
MSG_WELCOME_TITLE=Welcome
MSG_WELCOME_SUBTITLE=Validate your exclusive access by entering your CPF below
MSG_CPF_LABEL=Taxpayer ID
MSG_SUCCESS_MSG=Welcome to the Club!
MSG_NOT_FOUND=ID not found in our database
```

---

### Exemplo 3: Customizar Validação

**config.yaml:**
```yaml
behavior:
  condo_id_required: true
  default_condo_id: "condo_123"
  redirect_url_on_success: "https://meusite.com/dashboard"
  auto_close_modal_seconds: 5

validation:
  max_retries: 5
  retry_delay_ms: [500, 1000, 2000, 4000, 8000]
```

---

### Exemplo 4: Customizar Integrações

**config.yaml:**
```yaml
integrations:
  name_integration:
    enabled: true
    type: "custom"
    url: "https://minha-api.com/validar"
    
  partner_integration:
    enabled: true
    type: "meu_partner"
    url: "https://partner.com/registrar"
```

## 🚀 Endpoints Disponíveis

### `GET /config` - Retorna Configuração

Retorna toda a configuração em JSON que o frontend pode ler:

```bash
curl http://localhost:8080/config
```

**Resposta:**
```json
{
  "branding": {
    "app_name": "VIP Lounge",
    "company_name": "VIP Lounge",
    "theme_color": "4f46e5",
    "secondary_color": "8b5cf6"
  },
  "messages": {
    "welcome_title": "Bem-vindo",
    "success_message": "Bem-vindo ao Clube!"
    // ... todos os textos
  },
  "behavior": {
    "enable_debug_panel": true,
    "language": "pt-BR",
    "show_user_id_in_modal": true
  }
}
```

### `POST /v1/validate` - Validar CPF

Funciona normalmente, mas agora usa mensagens da config:

```bash
curl -X POST http://localhost:8080/v1/validate \
  -H "Content-Type: application/json" \
  -d '{"cpf": "00933733844", "condo_id": "13"}'
```

## 📁 Prioridade de Configuração

**De menor para maior prioridade:**

1. **Defaults hardcoded** em `config.go`
2. **`config.yaml`** (se existir)
3. **Variáveis de Ambiente** (`.env`)

Exemplo:
- Padrão: `app_name = "VIP Lounge"`
- YAML sobrescreve: `app_name = "Meu App"`
- ENV sobrescreve tudo: `APP_NAME=App Final`

## 🎨 Customização Visual

### Cores CSS Dinâmicas

O frontend aplica as cores de tema automaticamente:

```html
<!-- As cores da config são injetadas como CSS -->
<style>
  :root {
    --theme-color: #4f46e5;
    --secondary-color: #8b5cf6;
  }
</style>
```

### Buttons

Todos os botões usam `--theme-color`:

```html
<button style="background: var(--theme-color)">Validar</button>
```

## 📊 Multi-Cliente / Multi-Tenancy

Para servir múltiplos clientes com configs diferentes:

### Opção 1: Usar variáveis de ambiente por cliente

```bash
# Cliente 1
APP_NAME="Cliente 1" \
COMPANY_EMAIL="cliente1@email.com" \
PORT=8081 \
go run cmd/server/main.go

# Cliente 2
APP_NAME="Cliente 2" \
COMPANY_EMAIL="cliente2@email.com" \
PORT=8082 \
go run cmd/server/main.go
```

### Opção 2: Usar diferentes arquivos YAML

```bash
go run cmd/server/main.go --config=config-cliente-1.yaml
go run cmd/server/main.go --config=config-cliente-2.yaml
```

*(Nota: Isso requer adicionar flag de CLI em `main.go`)*

## 🔒 Segurança

### Debug Panel

Desabilitar em produção:

```bash
ENABLE_DEBUG=false
```

### CORS

Restringir origens em produção:

```bash
CORS_ORIGINS="https://meusite.com,https://app.meusite.com"
```

### HTTPS Obrigatório

```bash
REQUIRE_HTTPS=true
```

## 📝 Checklist para Novo Cliente

1. ✅ Criar `config.yaml` com branding do cliente
2. ✅ Configurar variáveis de ambiente (`.env`)
3. ✅ Customizar mensagens para o idioma/cultura
4. ✅ Configurar integrações (APIs específicas)
5. ✅ Testar endpoint `GET /config`
6. ✅ Testar fluxo completo no browser
7. ✅ Desabilitar debug panel em produção
8. ✅ Deploy com Cloud Run/Docker

## 🧪 Testando Customizações

### 1. Modificar `config.yaml` e reiniciar

```bash
# Editar config.yaml
# Mudar: app_name = "Novo Nome"

# Reiniciar servidor
go run cmd/server/main.go
```

### 2. Verificar configuração retornada

```bash
curl http://localhost:8080/config | jq '.branding.app_name'
# Output: "Novo Nome"
```

### 3. Verificar no browser

Abrir http://localhost:8080 e conferir se:
- Título da página mudou
- Mensagens estão corretas
- Cores aplicadas corretamente

## 🆚 Antes vs Depois (Agnóstico)

**ANTES:**
- Nome "VIP Lounge" hardcoded em 5 lugares
- Mensagens hardcoded em português
- Cores hardcoded em Tailwind
- Modificar código para novo cliente
- Deploy novo para cada cliente

**DEPOIS:**
- Nome configurável em `config.yaml`
- Mensagens via variáveis de ambiente
- Cores dinâmicas via CSS variables
- Sem modificação de código
- Mesmo binário para múltiplos clientes

## 🤔 FAQs

**P: Como adicionar novo campo de configuração?**
R: 
1. Adicionar struct em `internal/config/config.go`
2. Adicionar campo em `config.yaml`
3. Ler em `loadFromEnv()` se necessário
4. Usar em `handler` ou `service`

**P: Preciso mudar o banco de dados?**
R: Configurar `DB_TYPE` em `.env`:
```bash
DB_TYPE=postgres  # ou mongodb
```

**P: Como fazer A/B testing?**
R: Servir configs diferentes por query param:
```javascript
const clientId = new URLSearchParams(location.search).get('client');
fetch(`/config?client=${clientId}`)
```

**P: Frontend não está carregando as mensagens?**
R: Verificar console do browser (F12) e verificar `/config` responde JSON válido.

---

**🎉 Parabéns! Seu sistema agora é verdadeiramente agnóstico e pronto para múltiplos clientes!**
