package service

import (
	"context"
	"time"

	"github.com/viplounge/platform/internal/domain"
)

type ValidationService struct {
	repo      domain.LeadRepository
	validator domain.BenefValidator
	partner   domain.PartnerService
}

func NewValidationService(repo domain.LeadRepository, validator domain.BenefValidator, partner domain.PartnerService) *ValidationService {
	return &ValidationService{
		repo:      repo,
		validator: validator,
		partner:   partner,
	}
}

func (s *ValidationService) ValidateAndSave(ctx context.Context, req domain.ValidationRequest) (*domain.ValidationResponse, error) {
	// 1. Chama o Adapter (Benef API)
	isValid, leadData, err := s.validator.ValidateMember(ctx, req.CondoID, req.CPF)
	if err != nil {
		return nil, err
	}

	response := &domain.ValidationResponse{
		Valid: isValid,
	}

	// 2. Prepara o Lead para persistência
	leadToSave := domain.Lead{
		CPF:       req.CPF,
		CondoID:   req.CondoID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Origin:    "landing_page",
	}

	if isValid && leadData != nil {
		leadToSave.Status = domain.StatusApproved
		leadToSave.Name = leadData.Name
		leadToSave.Email = leadData.Email
		leadToSave.Phone = leadData.Phone
		leadToSave.SuperlogicaFound = leadData.SuperlogicaFound
		leadToSave.SuperlogicaResponseMs = leadData.SuperlogicaResponseMs
		leadToSave.RedeParceriasStatus = domain.PartnerStatusPending
		leadToSave.Metadata = leadData.Metadata
		
		// 3. Integração Rede Parcerias
		if err := s.partner.RegisterUser(ctx, &leadToSave); err != nil {
			// Log do erro mas não bloqueia o fluxo
			// O usuário vê sucesso pois foi validado na Superlogica
			leadToSave.RedeParceriasStatus = domain.PartnerStatusRetryPending
			response.Status = domain.ResponseStatusError
			response.Message = "Erro ao cadastrar no clube. Tente novamente."
		} else {
			// Se não retornou erro, pode ser sucesso ou já existia (422)
			leadToSave.RedeParceriasStatus = domain.PartnerStatusRegistered
			
			// Verificar se é um caso de "já existe" (422)
			if leadToSave.RedeParceriasError == "USER_ALREADY_EXISTS (422)" {
				response.Status = domain.ResponseStatusAlreadyRegistered
				response.Message = "Você já está cadastrado em nosso clube de beneficiários!"
			} else {
				response.Status = domain.ResponseStatusSuccess
				response.Message = "Bem-vindo ao Clube!"
			}
		}

		response.Name = leadData.Name
	} else {
		leadToSave.Status = domain.StatusRejected
		leadToSave.SuperlogicaFound = false
		response.Status = domain.ResponseStatusNotFound
		response.Message = "Condomínio não participante ou CPF não encontrado."
	}

	// 4. Salva no Firestore (se repo estiver disponível)
	if s.repo != nil {
		_ = s.repo.Save(ctx, leadToSave)
	}
	
	// 5. Adiciona user_id à resposta se disponível
	if leadToSave.RedeParceriasUserID != "" {
		response.UserID = leadToSave.RedeParceriasUserID
	}
	// Em produção, usaríamos log estruturado (slog/zap) aqui

	return response, nil
}


