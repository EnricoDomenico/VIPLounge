
package benef

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/viplounge/platform/internal/domain"
)

// SuperlogicaAdapter implementa a interface domain.BenefValidator
type SuperlogicaAdapter struct {
	apiURL      string
	appToken    string
	accessToken string
	// condoID removido pois agora é dinâmico
}

func NewBenefAdapter() *SuperlogicaAdapter {
	// Base URL sem o endpoint final, pois usaremos múltiplos endpoints
	url := os.Getenv("SUPERLOGICA_URL")
	if url == "" {
		url = "https://api.superlogica.net/v2/condor"
	}
	
	appToken := os.Getenv("SUPERLOGICA_APP_TOKEN")
	if appToken == "" {
		log.Printf("[WARN] SUPERLOGICA_APP_TOKEN não definido! Configure via variável de ambiente no Render.")
		// Fallback apenas para desenvolvimento local
		appToken = "74539367-69b7-432a-934f-8d9050bade0c"
	}

	accessToken := os.Getenv("SUPERLOGICA_ACCESS_TOKEN")
	if accessToken == "" {
		log.Printf("[WARN] SUPERLOGICA_ACCESS_TOKEN não definido! Configure via variável de ambiente no Render.")
		// Fallback apenas para desenvolvimento local
		accessToken = "d769811d-2d05-4640-b756-b2bae62318cd"
	}

	log.Printf("[SUPERLOGICA] Inicializado - URL: %s", url)

	return &SuperlogicaAdapter{
		apiURL:      url,
		appToken:    appToken,
		accessToken: accessToken,
	}
}

// Response structs
type UnitResponse []struct {
	IDUnidade            string `json:"id_unidade_uni"`
	IDCondominio         string `json:"id_condominio_cond"` // ID do condomínio retornado pela API
	NomeProprietario     string `json:"nome_proprietario"`
	EmailProprietario    string `json:"email_proprietario"`
	CelularProprietario  string `json:"celular_proprietario"`
	TelefoneProprietario string `json:"telefone_proprietario"`
	CPFProprietario      string `json:"cpf_proprietario"`
}

func (s *SuperlogicaAdapter) ValidateMember(ctx context.Context, condoID string, cpf string) (bool, *domain.Lead, error) {
	// BUSCA GLOBAL: Usar idCondominio=-1 para buscar em todos os condomínios
	// Se condoID for vazio ou "-1", a Superlógica fará a busca global
	// Se condoID for específico (ex: "4", "558"), busca apenas naquele condomínio
	
	if condoID == "" {
		condoID = "-1" // Busca global por padrão
	}
	
	log.Printf("[BENEF] ValidateMember - CondoID: %s, CPF: %s", condoID, cpf)
	
	// Faz a busca (global ou específica)
	found, lead, err := s.checkUnit(ctx, condoID, cpf)
	
	if err != nil {
		log.Printf("[BENEF] Erro na busca: %v", err)
		return false, nil, err
	}
	
	if found {
		log.Printf("[BENEF] Morador encontrado! Condomínio real: %s, Nome: %s", lead.CondoID, lead.Name)
		return true, lead, nil
	}
	
	log.Printf("[BENEF] Morador não encontrado para CPF: %s", cpf)
	return false, nil, nil
}

func (s *SuperlogicaAdapter) checkUnit(ctx context.Context, id string, cpf string) (bool, *domain.Lead, error) {
	startTime := time.Now()
	
	endpoint := fmt.Sprintf("%s/unidades/index", s.apiURL)
	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	q := req.URL.Query()
	q.Add("idCondominio", id)
	q.Add("pesquisa", cpf)
	q.Add("itensPorPagina", "50") // Aumentado conforme solicitado
	q.Add("exibirDadosDosContatos", "1")
	req.URL.RawQuery = q.Encode()

	s.addHeaders(req)
	
	log.Printf("[BENEF] Chamando API: %s?idCondominio=%s&pesquisa=%s", endpoint, id, cpf)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var data UnitResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false, nil, err
	}

	if len(data) > 0 {
		unit := data[0]
		phone := unit.CelularProprietario
		if phone == "" {
			phone = unit.TelefoneProprietario
		}
		
		responseTime := time.Since(startTime).Milliseconds()
		
		// CAPTURA O ID REAL DO CONDOMÍNIO DA RESPOSTA DA API
		// Quando idCondominio=-1, a API retorna o id_condominio_cond correto
		realCondoID := unit.IDCondominio
		if realCondoID == "" {
			realCondoID = id // Fallback: usa o ID da busca se não vier na resposta
		}
		
		log.Printf("[BENEF] Morador encontrado! Condomínio Real: %s (busca com: %s)", realCondoID, id)

		return true, &domain.Lead{
			CPF:                   cpf,
			CondoID:               realCondoID, // USA O ID REAL retornado pela API
			Name:                  unit.NomeProprietario,
			Email:                 unit.EmailProprietario,
			Phone:                 phone,
			Status:                domain.StatusApproved,
			Origin:                "superlogica_api",
			SuperlogicaFound:      true,
			SuperlogicaResponseMs: responseTime,
			RedeParceriasStatus:   domain.PartnerStatusPending,
			RedeParceriasAttempts: 0,
		}, nil
	}

	return false, nil, nil
}

func (s *SuperlogicaAdapter) addHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("app_token", s.appToken)
	req.Header.Add("access_token", s.accessToken)
}
