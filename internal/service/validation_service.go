package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/viplounge/platform/internal/config"
	"github.com/viplounge/platform/internal/domain"
)

type ValidationService struct {
	repo      domain.LeadRepository
	validator domain.BenefValidator
	partner   domain.PartnerService
	cfg       *config.Config
}

func NewValidationService(repo domain.LeadRepository, validator domain.BenefValidator, partner domain.PartnerService, cfg *config.Config) *ValidationService {
	if cfg == nil {
		cfg = config.Get()
	}
	return &ValidationService{
		repo:      repo,
		validator: validator,
		partner:   partner,
		cfg:       cfg,
	}
}

// ValidateAndSave implementa a ÁRVORE DE DECISÃO completa
func (s *ValidationService) ValidateAndSave(ctx context.Context, req domain.ValidationRequest) (*domain.ValidationResponse, error) {
	response := &domain.ValidationResponse{}
	
	// Preparar lead para tracking
	lead := domain.Lead{
		CPF:       req.CPF,
		CondoID:   req.CondoID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Origin:    "landing_page",
	}

	// ===== PASSO 1: Verificar na Superlógica =====
	log.Printf("[VALIDAÇÃO] Verificando CPF %s na Superlógica...", maskCPF(req.CPF))
	
	existsInSuperlogica, superlogicaData, superlogicaErr := s.validator.ValidateMember(ctx, req.CondoID, req.CPF)
	
	if superlogicaErr != nil {
		log.Printf("[ERRO] Falha na Superlógica: %v", superlogicaErr)
		// Não bloquear fluxo, continuar verificação na Rede Parcerias
	}

	if existsInSuperlogica && superlogicaData != nil {
		lead.Name = superlogicaData.Name
		lead.Email = superlogicaData.Email
		lead.Phone = superlogicaData.Phone
		lead.SuperlogicaFound = true
		lead.SuperlogicaResponseMs = superlogicaData.SuperlogicaResponseMs
	}

	// ===== PASSO 2: Verificar na Rede Parcerias =====
	log.Printf("[VALIDAÇÃO] Verificando CPF %s na Rede Parcerias...", maskCPF(req.CPF))
	
	var existsInPartner bool
	var partnerUser *domain.PartnerUser
	
	partnerUser, partnerErr := s.partner.FindUserByCPF(ctx, req.CPF)
	if partnerErr != nil {
		log.Printf("[WARN] Erro ao verificar Rede Parcerias: %v", partnerErr)
		// Continuar fluxo assumindo que não existe
	}
	existsInPartner = partnerUser != nil

	// ===== PASSO 3: ÁRVORE DE DECISÃO =====
	log.Printf("[DECISÃO] Superlógica=%v, RedeParcerias=%v", existsInSuperlogica, existsInPartner)

	switch {
	// CENÁRIO 1: Na Superlógica + NÃO na Rede Parcerias → NOVO USUÁRIO
	case existsInSuperlogica && !existsInPartner:
		log.Printf("[CENÁRIO] NOVO USUÁRIO - Cadastrar na Rede Parcerias")
		response = s.handleNewUser(ctx, &lead)

	// CENÁRIO 2: Na Superlógica + JÁ na Rede Parcerias → USUÁRIO EXISTENTE
	case existsInSuperlogica && existsInPartner:
		log.Printf("[CENÁRIO] USUÁRIO EXISTENTE - Gerar SSO")
		response = s.handleExistingUser(ctx, &lead, partnerUser)

	// CENÁRIO 3: NÃO na Superlógica + NA Rede Parcerias → REVOGAR ACESSO
	case !existsInSuperlogica && existsInPartner:
		log.Printf("[CENÁRIO] REVOGAR ACESSO - Usuário não está mais na Superlógica")
		response = s.handleRevokedUser(ctx, &lead, partnerUser)

	// CENÁRIO 4: NÃO existe em nenhum sistema → MARKETING
	default:
		log.Printf("[CENÁRIO] NÃO ENCONTRADO - Exibir marketing")
		response = s.handleNotFound(ctx, &lead)
	}

	// ===== PASSO 4: Salvar lead para analytics =====
	if s.repo != nil {
		if err := s.repo.Save(ctx, lead); err != nil {
			log.Printf("[WARN] Erro ao salvar lead: %v", err)
		}
	}

	return response, nil
}

// handleNewUser - CPF na Superlógica, NÃO na Rede Parcerias
// Ação: Cadastrar e gerar SSO
func (s *ValidationService) handleNewUser(ctx context.Context, lead *domain.Lead) *domain.ValidationResponse {
	response := &domain.ValidationResponse{
		Valid:              true,
		Scenario:          domain.ScenarioNewUser,
		Name:               lead.Name,
		Email:              lead.Email,
		Message:            "Você tem direito ao Clube de Benefícios!",
		ShowActivateButton: true,
	}

	// Cadastrar na Rede Parcerias
	if err := s.partner.RegisterUser(ctx, lead); err != nil {
		log.Printf("[ERRO] Falha ao cadastrar na Rede Parcerias: %v", err)
		response.Message = "Você tem direito ao benefício! Clique para ativar sua conta."
		return response
	}

	// Se cadastrou com sucesso, gerar SSO
	if lead.RedeParceriasUserID != "" {
		sso, err := s.partner.GetSSOToken(ctx, lead.RedeParceriasUserID)
		if err != nil {
			log.Printf("[WARN] Falha ao gerar SSO: %v", err)
		} else {
			response.SSOToken = sso.Token
			response.RedirectURL = sso.Redirect
			response.UserID = lead.RedeParceriasUserID
			response.ShowActivateButton = false
			response.Message = "Conta ativada com sucesso! Redirecionando..."
		}
	}

	lead.Status = domain.StatusApproved
	lead.RedeParceriasStatus = domain.PartnerStatusRegistered

	return response
}

// handleExistingUser - CPF na Superlógica E na Rede Parcerias
// Ação: Apenas gerar SSO para login
func (s *ValidationService) handleExistingUser(ctx context.Context, lead *domain.Lead, partnerUser *domain.PartnerUser) *domain.ValidationResponse {
	response := &domain.ValidationResponse{
		Valid:    true,
		Scenario: domain.ScenarioExistingUser,
		Name:     lead.Name,
		Email:    lead.Email,
		UserID:   partnerUser.ID,
		Message:  fmt.Sprintf("Bem-vindo de volta, %s!", firstName(lead.Name)),
	}

	// Gerar SSO para login automático
	sso, err := s.partner.GetSSOToken(ctx, partnerUser.ID)
	if err != nil {
		log.Printf("[WARN] Falha ao gerar SSO: %v", err)
		response.Message = "Sua conta está ativa! Acesse o Clube de Benefícios."
	} else {
		response.SSOToken = sso.Token
		response.RedirectURL = sso.Redirect
		response.Message = fmt.Sprintf("Bem-vindo de volta, %s! Redirecionando...", firstName(lead.Name))
	}

	lead.Status = domain.StatusApproved
	lead.RedeParceriasStatus = domain.PartnerStatusRegistered
	lead.RedeParceriasUserID = partnerUser.ID

	return response
}

// handleRevokedUser - CPF NÃO na Superlógica, mas NA Rede Parcerias
// Ação: Revogar acesso e mostrar marketing
func (s *ValidationService) handleRevokedUser(ctx context.Context, lead *domain.Lead, partnerUser *domain.PartnerUser) *domain.ValidationResponse {
	response := &domain.ValidationResponse{
		Valid:          false,
		Scenario:       domain.ScenarioRevokedUser,
		Message:        "Seu acesso ao Clube foi revogado pois você não consta mais como condômino.",
		ShowMarketing2: true,
	}

	// Desativar na Rede Parcerias
	if err := s.partner.DeleteUser(ctx, partnerUser.ID); err != nil {
		log.Printf("[WARN] Falha ao revogar acesso: %v", err)
	} else {
		log.Printf("[INFO] Acesso revogado com sucesso para user %s", partnerUser.ID)
	}

	lead.Status = domain.StatusRejected
	lead.RedeParceriasStatus = domain.PartnerStatusRevoked

	return response
}

// handleNotFound - CPF não existe em nenhum sistema
// Ação: Mostrar marketing para atrair novo cliente
func (s *ValidationService) handleNotFound(ctx context.Context, lead *domain.Lead) *domain.ValidationResponse {
	response := &domain.ValidationResponse{
		Valid:          false,
		Scenario:       domain.ScenarioNotFound,
		Message:        "CPF não encontrado. Seja um condômino para ter acesso aos benefícios exclusivos!",
		ShowMarketing1: true,
	}

	lead.Status = domain.StatusRejected
	lead.SuperlogicaFound = false

	return response
}

// Helpers
func maskCPF(cpf string) string {
	if len(cpf) < 6 {
		return "***"
	}
	return cpf[:3] + ".***.***-" + cpf[len(cpf)-2:]
}

func firstName(fullName string) string {
	if fullName == "" {
		return "usuário"
	}
	for i, c := range fullName {
		if c == ' ' {
			return fullName[:i]
		}
	}
	return fullName
}
