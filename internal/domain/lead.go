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

// Status da Integracao com Rede Parcerias
const (
	PartnerStatusPending      = "PENDING"
	PartnerStatusRegistered   = "REGISTERED"
	PartnerStatusRetryPending = "RETRY_PENDING"
	PartnerStatusFailed       = "FAILED"
)

// Status de Resposta para o Frontend
const (
	ResponseStatusSuccess           = "success"           // CPF validado e cadastrado
	ResponseStatusAlreadyRegistered = "already_registered" // CPF já estava cadastrado (422)
	ResponseStatusNotFound          = "not_found"         // CPF não encontrado na Superlogica
	ResponseStatusError             = "error"             // Erro geral
)

// Lead representa um condômino/usuário validado ou tentado
type Lead struct {
	// IDs
	CPF       string                 `json:"cpf" firestore:"cpf"`
	CondoID   string                 `json:"condo_id" firestore:"condo_id"`
	
	// Dados Pessoais
	Name      string                 `json:"name,omitempty" firestore:"name,omitempty"`
	Email     string                 `json:"email,omitempty" firestore:"email,omitempty"`
	Phone     string                 `json:"phone,omitempty" firestore:"phone,omitempty"`
	
	// Status
	Status    string                 `json:"status" firestore:"status"`
	Origin    string                 `json:"origin" firestore:"origin"`
	
	// Superlogica Metrics
	SuperlogicaFound      bool       `json:"superlogica_found" firestore:"superlogica_found"`
	SuperlogicaResponseMs int64      `json:"superlogica_response_ms" firestore:"superlogica_response_ms"`
	
	// Rede Parcerias Integration
	RedeParceriasStatus   string     `json:"rede_parcerias_status" firestore:"rede_parcerias_status"`
	RedeParceriasUserID   string     `json:"rede_parcerias_user_id" firestore:"rede_parcerias_user_id"`
	RedeParceriasAttempts int        `json:"rede_parcerias_attempts" firestore:"rede_parcerias_attempts"`
	RedeParceriasError    string     `json:"rede_parcerias_error" firestore:"rede_parcerias_error,omitempty"`
	RedeParceriasResponseMs int64    `json:"rede_parcerias_response_ms" firestore:"rede_parcerias_response_ms"`
	
	// Auditoria
	CreatedAt time.Time              `json:"created_at" firestore:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" firestore:"updated_at"`
	
	// Metadata
	Metadata  map[string]interface{} `json:"metadata,omitempty" firestore:"metadata,omitempty"`
}

// ValidationRequest é o payload recebido da Landing Page
type ValidationRequest struct {
	CPF     string `json:"cpf"`
	CondoID string `json:"condo_id"`
}

// ValidationResponse é a resposta para o Frontend
type ValidationResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message"`
	Name    string `json:"name,omitempty"`
	Status  string `json:"status,omitempty"`
	UserID  string `json:"user_id,omitempty"`
}

// LeadRepository define como salvamos os leads (Porta de Saída)
type LeadRepository interface {
	Save(ctx context.Context, lead Lead) error
}

// BenefValidator define o contrato para validar um CPF em uma API externa (Porta de Saída)
type BenefValidator interface {
	ValidateMember(ctx context.Context, condoID string, cpf string) (bool, *Lead, error)
}

// PartnerService define o contrato para cadastrar o usuário no clube de benefícios
type PartnerService interface {
	RegisterUser(ctx context.Context, lead *Lead) error
}
