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
// Ação: Retornar pending_email_confirmation para validação em duas etapas
func (s *ValidationService) handleNewUser(ctx context.Context, lead *domain.Lead) *domain.ValidationResponse {
	response := &domain.ValidationResponse{
		Valid:              true,
		Scenario:           domain.ScenarioPendingEmailConfirmation,
		Name:               lead.Name,
		EmailHint:          maskEmail(lead.Email),
		Message:            "Você tem direito ao Clube de Benefícios!",
		ShowActivateButton: false,
	}

	// NÃO gerar SSO neste momento - aguardar confirmação de e-mail
	// O e-mail real será mantido no backend para validação posterior
	log.Printf("[SEGURANÇA] Nova validação em duas etapas - Email mascarado: %s", response.EmailHint)

	lead.Status = domain.StatusPending // Status pendente até confirmar e-mail
	lead.RedeParceriasStatus = domain.PartnerStatusPending

	return response
}

// handleExistingUser - CPF na Superlógica E na Rede Parcerias
// Ação: Retornar pending_email_confirmation para validação em duas etapas
func (s *ValidationService) handleExistingUser(ctx context.Context, lead *domain.Lead, partnerUser *domain.PartnerUser) *domain.ValidationResponse {
	response := &domain.ValidationResponse{
		Valid:     true,
		Scenario:  domain.ScenarioPendingEmailConfirmation,
		Name:      lead.Name,
		EmailHint: maskEmail(lead.Email),
		UserID:    partnerUser.ID,
		Message:   fmt.Sprintf("Bem-vindo de volta, %s!", firstName(lead.Name)),
	}

	// NÃO gerar SSO neste momento - aguardar confirmação de e-mail
	log.Printf("[SEGURANÇA] Usuário existente - Email mascarado: %s", response.EmailHint)

	lead.Status = domain.StatusPending // Status pendente até confirmar e-mail
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

// ConfirmEmailAndActivate valida o e-mail fornecido e, se correto, gera o SSO
func (s *ValidationService) ConfirmEmailAndActivate(ctx context.Context, req domain.EmailConfirmationRequest) (*domain.ValidationResponse, error) {
	response := &domain.ValidationResponse{}
	
	// Preparar lead para tracking
	lead := domain.Lead{
		CPF:       req.CPF,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Origin:    "email_confirmation",
	}

	// ===== PASSO 1: Verificar na Superlógica novamente =====
	log.Printf("[CONFIRMAÇÃO] Verificando CPF %s na Superlógica...", maskCPF(req.CPF))
	
	existsInSuperlogica, superlogicaData, superlogicaErr := s.validator.ValidateMember(ctx, s.cfg.Behavior.DefaultCondoID, req.CPF)
	
	if superlogicaErr != nil || !existsInSuperlogica || superlogicaData == nil {
		log.Printf("[ERRO] CPF não encontrado na confirmação: %v", superlogicaErr)
		response.Valid = false
		response.Scenario = domain.ScenarioError
		response.Message = "CPF não encontrado. Por favor, inicie o processo novamente."
		return response, nil
	}

	lead.Name = superlogicaData.Name
	lead.Email = superlogicaData.Email
	lead.Phone = superlogicaData.Phone
	lead.SuperlogicaFound = true

	// ===== PASSO 2: Validar e-mail digitado =====
	log.Printf("[VALIDAÇÃO] Comparando e-mail fornecido com cadastrado...")
	
	// Normalizar e-mails para comparação (lowercase e trim)
	providedEmail := normalizeEmail(req.Email)
	registeredEmail := normalizeEmail(superlogicaData.Email)
	
	if providedEmail != registeredEmail {
		log.Printf("[SEGURANÇA] E-mail incorreto! Esperado: %s, Recebido: %s", maskEmail(registeredEmail), maskEmail(providedEmail))
		response.Valid = false
		response.Scenario = domain.ScenarioError
		response.Message = "E-mail incorreto. Por favor, verifique o e-mail cadastrado."
		lead.Status = domain.StatusRejected
		
		// Salvar tentativa falha para auditoria
		if s.repo != nil {
			s.repo.Save(ctx, lead)
		}
		
		return response, nil
	}

	log.Printf("[SUCESSO] E-mail confirmado! Prosseguindo com ativação...")

	// ===== PASSO 3: Verificar se já existe na Rede Parcerias =====
	partnerUser, partnerErr := s.partner.FindUserByCPF(ctx, req.CPF)
	if partnerErr != nil {
		log.Printf("[WARN] Erro ao verificar Rede Parcerias: %v", partnerErr)
	}
	existsInPartner := partnerUser != nil

	// ===== PASSO 4: Ativar usuário =====
	if existsInPartner {
		// Usuário já existe - apenas gerar SSO
		log.Printf("[ATIVAÇÃO] Usuário já existe - gerando SSO")
		response = s.activateExistingUser(ctx, &lead, partnerUser)
	} else {
		// Novo usuário - cadastrar e gerar SSO
		log.Printf("[ATIVAÇÃO] Novo usuário - cadastrando e gerando SSO")
		response = s.activateNewUser(ctx, &lead)
	}

	// ===== PASSO 5: Salvar lead para analytics =====
	if s.repo != nil {
		if err := s.repo.Save(ctx, lead); err != nil {
			log.Printf("[WARN] Erro ao salvar lead: %v", err)
		}
	}

	return response, nil
}

// activateNewUser cadastra novo usuário e gera SSO
func (s *ValidationService) activateNewUser(ctx context.Context, lead *domain.Lead) *domain.ValidationResponse {
	response := &domain.ValidationResponse{
		Valid:    true,
		Scenario: domain.ScenarioNewUser,
		Name:     lead.Name,
		Email:    lead.Email,
		Message:  "Conta ativada com sucesso!",
	}

	// Cadastrar e gerar SSO
	sso, err := s.partner.RegisterAndGetSSO(ctx, lead)
	if err != nil {
		log.Printf("[ERRO] Falha no RegisterAndGetSSO: %v", err)
		
		// Fallback: tentar só o SSO se tiver email
		if lead.Email != "" {
			log.Printf("[RETRY] Tentando gerar SSO usando email: %s", lead.Email)
			sso, err = s.partner.GetSSOToken(ctx, lead.Email)
			if err != nil {
				log.Printf("[ERRO] Fallback SSO também falhou: %v", err)
				response.Message = "Você tem direito ao benefício mas houve um erro. Tente novamente."
				return response
			}
		} else {
			response.Message = "Você tem direito ao benefício mas houve um erro. Tente novamente."
			return response
		}
	}

	// Sucesso
	if sso != nil && sso.Redirect != "" {
		response.SSOToken = sso.Token
		response.RedirectURL = sso.Redirect
		response.UserID = lead.RedeParceriasUserID
		response.Message = "Conta ativada com sucesso! Redirecionando para o Clube de Benefícios..."
		log.Printf("[SUCESSO] SSO gerado! Redirect: %s", sso.Redirect)
	}

	lead.Status = domain.StatusApproved
	lead.RedeParceriasStatus = domain.PartnerStatusRegistered

	return response
}

// activateExistingUser gera SSO para usuário existente
func (s *ValidationService) activateExistingUser(ctx context.Context, lead *domain.Lead, partnerUser *domain.PartnerUser) *domain.ValidationResponse {
	response := &domain.ValidationResponse{
		Valid:    true,
		Scenario: domain.ScenarioExistingUser,
		Name:     lead.Name,
		Email:    lead.Email,
		UserID:   partnerUser.ID,
		Message:  fmt.Sprintf("Bem-vindo de volta, %s!", firstName(lead.Name)),
	}

	// Gerar SSO
	ssoIdentifier := partnerUser.Email
	if ssoIdentifier == "" {
		ssoIdentifier = partnerUser.ID
	}

	sso, err := s.partner.GetSSOToken(ctx, ssoIdentifier)
	if err != nil {
		log.Printf("[WARN] Falha ao gerar SSO com %s: %v", ssoIdentifier, err)
		
		// Fallback
		if partnerUser.Email != "" && ssoIdentifier == partnerUser.ID {
			sso, err = s.partner.GetSSOToken(ctx, partnerUser.Email)
		} else if partnerUser.ID != "" {
			sso, err = s.partner.GetSSOToken(ctx, partnerUser.ID)
		}
		
		if err != nil {
			log.Printf("[ERRO] SSO fallback também falhou: %v", err)
			response.Message = "Houve um erro ao gerar seu acesso. Tente novamente."
			return response
		}
	}

	// Sucesso
	if sso != nil && sso.Redirect != "" {
		response.SSOToken = sso.Token
		response.RedirectURL = sso.Redirect
		response.Message = fmt.Sprintf("Bem-vindo de volta, %s! Redirecionando...", firstName(lead.Name))
		log.Printf("[SUCESSO] SSO gerado para usuário existente! Redirect: %s", sso.Redirect)
	}

	lead.Status = domain.StatusApproved
	lead.RedeParceriasStatus = domain.PartnerStatusRegistered
	lead.RedeParceriasUserID = partnerUser.ID

	return response
}

// normalizeEmail normaliza um e-mail para comparação
func normalizeEmail(email string) string {
	// Converter para lowercase e remover espaços
	normalized := ""
	for _, c := range email {
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			if c >= 'A' && c <= 'Z' {
				normalized += string(c + 32) // Converter para lowercase
			} else {
				normalized += string(c)
			}
		}
	}
	return normalized
}

// Helpers
func maskCPF(cpf string) string {
	if len(cpf) < 6 {
		return "***"
	}
	return cpf[:3] + ".***.***-" + cpf[len(cpf)-2:]
}

func maskEmail(email string) string {
	if email == "" {
		return ""
	}
	
	// Encontrar a posição do @
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	
	if atIndex == -1 || atIndex < 3 {
		// Se não tem @ ou é muito curto, mascarar tudo
		return "*****@***"
	}
	
	// Pegar os 3 últimos caracteres antes do @
	localPart := email[:atIndex]
	domain := email[atIndex:]
	
	if len(localPart) <= 3 {
		// Se tem 3 ou menos caracteres, mostrar apenas o último
		return "*****" + string(localPart[len(localPart)-1]) + domain
	}
	
	// Mostrar apenas os 3 últimos caracteres antes do @
	lastThree := localPart[len(localPart)-3:]
	return "*****" + lastThree + domain
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
