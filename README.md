# VIP Lounge Platform

Uma plataforma agnóstica de validação e cadastro de usuários. O sistema integra validação de CPF através de qualquer API de validação, registra usuários em qualquer plataforma de parceiros e persiste dados em qualquer banco de dados.

## 🎯 Visão Geral

O VIP Lounge é uma plataforma **completamente agnóstica** que permite:
- ✅ Customizar branding sem modificar código
- ✅ Suportar múltiplos idiomas via configuração
- ✅ Integrar com diferentes APIs de validação
- ✅ Registrar usuários em diferentes plataformas parceiras
- ✅ Usar diferentes bancos de dados
- ✅ Servir múltiplos clientes com a mesma instância

**Integrações padrão:**
- **Superlogica API**: Validação de CPF e dados de condôminos
- **Rede Parcerias API**: Registro como beneficiário de clube
- **Google Firestore**: Persistência e auditoria
- **Google Cloud Logging**: Logs estruturados

## ⚙️ Stack Técnico

- **Backend**: Go 1.21+
- **Frontend**: HTML5 + Tailwind CSS + Vanilla JavaScript
- **Persistência**: Google Cloud Firestore
- **Logging**: Google Cloud Logging
- **Infraestrutura**: Docker + Cloud Run + Cloud Build

## 📋 Pré-requisitos

### Obrigatório

- **Go 1.21+** - [Baixar aqui](https://go.dev/dl/)
- **Git** - Para controle de versão

### Opcional

- **Conta Google Cloud** - Para Firestore e Logging (opcional para dev local)

## 🎨 Customização (Agnóstico)

A plataforma é **100% agnóstica** e pode ser customizada sem modificar código:

### Via `config.yaml`

```yaml
branding:
  app_name: "Seu App"
  company_name: "Sua Empresa"
  theme_color: "4f46e5"

messages:
  welcome_title: "Bem-vindo"
  success_message: "Sucesso!"

behavior:
  language: "pt-BR"
  enable_debug_panel: true
```

### Via Variáveis de Ambiente (`.env`)

```env
APP_NAME=Seu App Customizado
COMPANY_EMAIL=contato@empresa.com
MSG_SUCCESS_MSG=Parabéns! Você foi aprovado!
LANGUAGE=en-US
```

**[Ver Guia Completo de Agnóstico](docs/AGNOSTIC_GUIDE.md)** para mais detalhes.

## 🚀 Início Rápido

### 1. Clonar Repositório

```bash
git clone https://github.com/EnricoDomenico/VIPLounge.git
cd VIPLounge
```

### 2. Configurar Variáveis de Ambiente

Copie o arquivo de exemplo:

```bash
cp .env.example .env
```

Edite `.env` com suas credenciais e customizações:

```env
# Branding
APP_NAME=Seu App
COMPANY_NAME=Sua Empresa
THEME_COLOR=4f46e5

# Mensagens
MSG_WELCOME_TITLE=Bem-vindo
MSG_SUCCESS_MSG=Sucesso!

# APIs
SUPERLOGICA_APP_TOKEN=seu-token
SUPERLOGICA_ACCESS_TOKEN=seu-token
REDE_PARCERIAS_BEARER_TOKEN=seu-jwt

# Servidor
PORT=8080
```

**Nota:** As variáveis de ambiente sobrescrevem `config.yaml`.

### 3. Baixar Dependências

```bash
go mod download
go mod tidy
```

### 4. Executar Servidor

```bash
go run cmd/server/main.go
```

Você verá:
```
2026/01/14 03:12:51 Server starting on port 8080
```

### 5. Acessar Aplicação

Abra no navegador: **http://localhost:8080**

## 📱 Como Usar

### Fluxo de Validação

1. **Preencher CPF**: Digite o CPF do titular (ex: `00933733844`)
2. **Validar**: Clique em "Validar Acesso"
3. **Resultado**:
   - ✅ **Success**: CPF encontrado → Novo cadastro no clube
   - ℹ️ **Already Registered**: CPF já cadastrado → Mostra ID existente
   - ❌ **Not Found**: CPF não faz parte do grupo participante

### Debug Panel

Um painel de debug aparece no canto inferior direito com:
- Status da resposta
- ID do usuário criado
- JSON completo da resposta
- Botão para copiar dados

**Nota**: O debug panel será removido na versão de produção.

## 📁 Estrutura do Projeto

```
.
├── cmd/
│   └── server/
│       └── main.go                 # Entrada da aplicação
├── internal/
│   ├── adapter/
│   │   ├── benef/                 # Integração Superlogica
│   │   │   └── api_interface.go
│   │   └── redeparcerias/         # Integração Rede Parcerias
│   │       └── client.go
│   ├── domain/
│   │   └── lead.go                # Modelos de dados
│   ├── handler/
│   │   └── http.go                # Rotas HTTP
│   ├── logger/
│   │   └── cloud_logger.go        # Logging estruturado
│   ├── repository/
│   │   └── firestore.go           # Persistência
│   └── service/
│       └── validation_service.go  # Lógica de negócio
├── web/
│   └── index.html                 # Frontend
├── .env.example                   # Variáveis de ambiente (exemplo)
├── .gitignore                     # Configuração Git
├── go.mod                         # Dependências Go
├── cloudbuild.yaml               # Configuração Cloud Build
├── Dockerfile                    # Imagem Docker
└── README.md                     # Este arquivo
```

## 🔌 APIs Integradas

### Superlogica API

**Endpoint**: `GET /v2/condor/unidades/index`

```bash
curl -X GET "https://api.superlogica.net/v2/condor/unidades/index?idCondominio=-1&pesquisa=00933733844&exibirDadosDosContatos=1" \
  -H "app_token: YOUR_TOKEN" \
  -H "access_token: YOUR_TOKEN"
```

**Resposta**: CPF encontrado retorna dados do titular

### Rede Parcerias API

**Endpoint**: `POST /api-client/v1/users`

```bash
curl -X POST "https://api.staging.clubeparcerias.com.br/api-client/v1/users" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{
    "name": "Nome Completo",
    "email": "email@example.com",
    "cpf": "00933733844",
    "authorized": true
  }'
```

**Resposta**: Status 201 com ID do usuário criado

## 🔐 Segurança

### Variáveis Sensíveis

Todas as credenciais estão em `.env` e **NUNCA** são commitadas ao Git:

```gitignore
.env           # ← Ignorado no Git
.env.example   # ← Versão pública (exemplo)
```

### Melhorias de Segurança

- ✅ Tokens em variáveis de ambiente
- ✅ CORS restrito em produção
- ✅ HTTPS obrigatório em produção
- ✅ Validação de CPF no cliente e servidor
- ✅ Logs estruturados para auditoria
- ✅ Credenciais do Firestore via Google Cloud Secret Manager

### Para Produção

1. Use **Cloud Secret Manager** ao invés de `.env`
2. Configure **Cloud Armor** para proteção DDoS
3. Habilite **VPC Service Controls** para isolamento
4. Ative **Cloud Audit Logs** para compliance

## 📊 Banco de Dados - Firestore

Estrutura de coleções:

### `leads` Collection

```json
{
  "id": "{condoID}_{cpf}",
  "cpf": "00933733844",
  "condo_id": "13",
  "name": "Ailton Geraldo Júnior",
  "email": "tico.agj@gmail.com",
  "status": "APPROVED",
  "superlogica_found": true,
  "superlogica_response_ms": 1150,
  "rede_parcerias_status": "REGISTERED",
  "rede_parcerias_user_id": "a0d4fedf-1c6d-4cc8-8f42-cbe9cc961ec4",
  "rede_parcerias_response_ms": 1023,
  "created_at": "2026-01-14T03:12:51Z",
  "updated_at": "2026-01-14T03:12:51Z"
}
```

## 🧪 Testando Localmente

### Com Servidor Rodando

1. Abra: `http://localhost:8080`
2. CPF de teste: `00933733844`
3. Clique "Validar Acesso"
4. Verifique o resultado

### Usando cURL

```bash
curl -X POST http://localhost:8080/v1/validate \
  -H "Content-Type: application/json" \
  -d '{
    "cpf": "00933733844",
    "condo_id": "13"
  }'
```

**Resposta esperada**:
```json
{
  "valid": true,
  "status": "success",
  "message": "Bem-vindo ao Clube!",
  "name": "Ailton Geraldo Júnior",
  "user_id": "a0d4fedf-1c6d-4cc8-8f42-cbe9cc961ec4"
}
```

## 🐳 Docker

### Build da Imagem

```bash
docker build -t viplounge:latest .
```

### Rodar Container

```bash
docker run -p 8080:8080 \
  -e SUPERLOGICA_APP_TOKEN=seu-token \
  -e SUPERLOGICA_ACCESS_TOKEN=seu-token \
  -e REDE_PARCERIAS_BEARER_TOKEN=seu-jwt \
  -e GOOGLE_CLOUD_PROJECT=seu-projeto \
  viplounge:latest
```

## ☁️ Deploy no Google Cloud Run

### 1. Authenticate

```bash
gcloud auth login
gcloud config set project seu-projeto-id
```

### 2. Build e Push

```bash
gcloud builds submit --tag gcr.io/seu-projeto-id/viplounge
```

### 3. Deploy

```bash
gcloud run deploy viplounge \
  --image gcr.io/seu-projeto-id/viplounge \
  --platform managed \
  --region us-central1 \
  --set-env-vars SUPERLOGICA_APP_TOKEN=seu-token,SUPERLOGICA_ACCESS_TOKEN=seu-token,REDE_PARCERIAS_BEARER_TOKEN=seu-jwt,GOOGLE_CLOUD_PROJECT=seu-projeto
```

## 📝 Endpoints

### Frontend
- `GET /` - Aplicação web dinâmica

### API - Validação
- `GET /health` - Health check
- `POST /v1/validate` - Validar CPF e registrar usuário
- `GET /config` - Retorna configuração agnóstica (consumido pelo frontend)

## 🛠️ Desenvolvimento

### Dependências

```bash
github.com/go-chi/chi/v5          # Router HTTP
github.com/go-chi/cors            # CORS middleware
cloud.google.com/go/firestore    # Firestore SDK
cloud.google.com/go/logging      # Cloud Logging SDK
```

### Adicionar Dependências

```bash
go get github.com/seu-pacote
go mod tidy
```

## 🐛 Troubleshooting

### Porta 8080 já em uso

```bash
# Windows
netstat -ano | findstr ":8080"
taskkill /PID <PID> /F

# Linux/Mac
lsof -i :8080
kill -9 <PID>
```

### Firestore não conecta

Isso é esperado em dev local. Configure credenciais do GCP:

```bash
gcloud auth application-default login
```

### Credenciais incorretas

Verifique se `.env` tem valores corretos:
```bash
cat .env | grep -E "SUPERLOGICA|REDE_PARCERIAS"
```

## 📚 Documentação Adicional

- [Superlogica API Docs](https://www.superlogica.com/api/)
- [Google Firestore](https://cloud.google.com/firestore/docs)
- [Google Cloud Logging](https://cloud.google.com/logging/docs)
- [Cloud Run Documentation](https://cloud.google.com/run/docs)

## 🤝 Contribuindo

1. Fork o repositório
2. Crie uma branch: `git checkout -b feature/sua-feature`
3. Commit suas mudanças: `git commit -m 'Add some feature'`
4. Push: `git push origin feature/sua-feature`
5. Abra um Pull Request

## 📄 Licença

MIT License - veja [LICENSE](LICENSE) para detalhes

## 👤 Autor

Enrico Domenico

## 📞 Suporte

Para problemas ou sugestões:
- Abra uma [Issue](https://github.com/EnricoDomenico/VIPLounge/issues)
- Envie um email

---

**Desenvolvido com ❤️ para VIP Lounge Platform**
