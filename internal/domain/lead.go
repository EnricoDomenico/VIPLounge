package domain

import (
	"context"
	"time"
)

// Status do Lead
const (
	StatusPending  = "PENDING"
	StatusApproved = "APPROVED"
	StatusRejected = "REJECTED"
	StatusError    = "ERROR"
)

// Status da Integração com Rede Parcerias
const (
	PartnerStatusPending      = "PENDING"
	PartnerStatusRegistered   = "REGISTERED"
	PartnerStatusRetryPending = "RETRY_PENDING"
	PartnerStatusFailed       = "FAILED"
	PartnerStatusRevoked      = "REVOKED"
)

// Cenários da Árvore de Decisão
const (
	// CPF na Superlógica + NÃO na Rede Parceira → Cadastrar
	ScenarioNewUser = "new_user"
	
	// CPF na Superlógica + JÁ na Rede Parceira → Apenas gerar SSO
	ScenarioExistingUser = "existing_user"
	
	// CPF NÃO na Superlógica + NA Rede Parceira → Revogar acesso
	ScenarioRevokedUser = "revoked_user"
	
	// CPF NÃO existe em nenhum sistema
	ScenarioNotFound = "not_found"
	
	// Erro durante validação
	ScenarioError = "error"
)

// Lead representa um condômino/usuário validado ou tentado
type Lead struct {
	// IDs
	CPF     string `json:"cpf" firestore:"cpf"`
	CondoID string `json:"condo_id" firestore:"condo_id"`

	// Dados Pessoais
	Name  string `json:"name,omitempty" firestore:"name,omitempty"`
	Email string `json:"email,omitempty" firestore:"email,omitempty"`
	Phone string `json:"phone,omitempty" firestore:"phone,omitempty"`

	// Status
	Status string `json:"status" firestore:"status"`
	Origin string `json:"origin" firestore:"origin"`

	// Superlogica Metrics
	SuperlogicaFound      bool  `json:"superlogica_found" firestore:"superlogica_found"`
	SuperlogicaResponseMs int64 `json:"superlogica_response_ms" firestore:"superlogica_response_ms"`

	// Rede Parcerias Integration
	RedeParceriasStatus     string `json:"rede_parcerias_status" firestore:"rede_parcerias_status"`
	RedeParceriasUserID     string `json:"rede_parcerias_user_id" firestore:"rede_parcerias_user_id"`
	RedeParceriasAttempts   int    `json:"rede_parcerias_attempts" firestore:"rede_parcerias_attempts"`
	RedeParceriasError      string `json:"rede_parcerias_error" firestore:"rede_parcerias_error,omitempty"`
	RedeParceriasResponseMs int64  `json:"rede_parcerias_response_ms" firestore:"rede_parcerias_response_ms"`

	// Auditoria
	CreatedAt time.Time `json:"created_at" firestore:"created_at"`
	UpdatedAt time.Time `json:"updated_at" firestore:"updated_at"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}

// PartnerUser representa um usuário na Rede Parcerias
type PartnerUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CPF       string `json:"cpf"`
	Cellphone string `json:"cellphone"`
	Active    bool   `json:"active"`
}

// SSOToken representa o token para login automático
type SSOToken struct {
	Token    string `json:"token"`
	Redirect string `json:"redirect"`
}

// ValidationRequest é o payload recebido da Landing Page
type ValidationRequest struct {
	CPF     string `json:"cpf"`
	CondoID string `json:"condo_id"`
}

// ValidationResponse é a resposta completa para o Frontend
type ValidationResponse struct {
	// Status básico
	Valid   bool   `json:"valid"`
	Message string `json:"message"`

	// Cenário identificado
	Scenario string `json:"scenario"`

	// Dados do usuário (se encontrado)
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	UserID string `json:"user_id,omitempty"`

	// SSO para redirecionamento (se aplicável)
	SSOToken    string `json:"sso_token,omitempty"`
	RedirectURL string `json:"redirect_url,omitempty"`

	// Flags para o frontend
	ShowActivateButton bool `json:"show_activate_button,omitempty"`
	ShowMarketing1     bool `json:"show_marketing_1,omitempty"`
	ShowMarketing2     bool `json:"show_marketing_2,omitempty"`
}

// LeadRepository define como salvamos os leads (Porta de Saída)
type LeadRepository interface {
	Save(ctx context.Context, lead Lead) error
}

// BenefValidator define o contrato para validar um CPF na Superlógica
type BenefValidator interface {
	ValidateMember(ctx context.Context, condoID string, cpf string) (bool, *Lead, error)
}

// PartnerService define o contrato para o Clube de Benefícios
type PartnerService interface {
	// Verificar se usuário existe
	FindUserByCPF(ctx context.Context, cpf string) (*PartnerUser, error)

	// Cadastrar novo usuário (com authorized:true)
	RegisterUser(ctx context.Context, lead *Lead) error

	// Deletar/desativar usuário
	DeleteUser(ctx context.Context, userID string) error

	// Gerar token SSO para redirecionamento
	// O userIdentifier pode ser UUID ou EMAIL
	GetSSOToken(ctx context.Context, userIdentifier string) (*SSOToken, error)

	// RegisterAndGetSSO faz o fluxo completo:
	// 1. Cadastra usuário (com authorized:true)
	// 2. Gera SSO Token usando EMAIL
	// 3. Retorna URL de redirect para login automático
	RegisterAndGetSSO(ctx context.Context, lead *Lead) (*SSOToken, error)
}
