
package benef

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
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
	NomeProprietario     string `json:"nome_proprietario"`
	EmailProprietario    string `json:"email_proprietario"`
	CelularProprietario  string `json:"celular_proprietario"`
	TelefoneProprietario string `json:"telefone_proprietario"`
	CPFProprietario      string `json:"cpf_proprietario"`
}

type CondoResponse []struct {
	IDCondominio string `json:"id_condominio_cond"`
}

func (s *SuperlogicaAdapter) ValidateMember(ctx context.Context, condoID string, cpf string) (bool, *domain.Lead, error) {
	// 1. TENTATIVA RÁPIDA: Busca Global (ID -1)
	// Se nenhum ID específico foi solicitado (ou foi solicitado busca global), tenta o atalho
	if condoID == "" || condoID == "-1" {
		found, lead, _ := s.checkUnit(ctx, "-1", cpf)
		if found {
			return true, lead, nil
		}
		// Se não achou ou deu erro no -1, cai para o fallback abaixo (Varredura)
	} else {
		// Se foi passado um ID específico (ex: "4"), verifica apenas nele
		return s.checkUnit(ctx, condoID, cpf)
	}

	// 2. FALLBACK: Varredura em todos os condomínios
	ids, err := s.listCondos(ctx)
	if err != nil {
		return false, nil, fmt.Errorf("erro ao listar condominios para varredura: %w", err)
	}

	// Itera sobre todos os condomínios
	for _, id := range ids {
		found, lead, err := s.checkUnit(ctx, id, cpf)
		if err == nil && found {
			return true, lead, nil
		}
		// Log de debug poderia ser útil aqui: "Não achou no condominio X"
	}

	return false, nil, nil
}

func (s *SuperlogicaAdapter) listCondos(ctx context.Context) ([]string, error) {
	endpoint := fmt.Sprintf("%s/condominios/index", s.apiURL)
	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	
	q := req.URL.Query()
	q.Add("itensPorPagina", "100") // Limite razoável para teste
	req.URL.RawQuery = q.Encode()

	s.addHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Se falhar listar, fallback para varredura manual 1..50 (brute force safe)
		var bruteList []string
		for i := 1; i <= 50; i++ {
			bruteList = append(bruteList, strconv.Itoa(i))
		}
		return bruteList, nil
	}

	var data CondoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var ids []string
	for _, c := range data {
		ids = append(ids, c.IDCondominio)
	}
	
	// Fallback se lista vazia ou erro
	if len(ids) == 0 {
		// Gera lista de 1 a 50 para varredura bruta garantida
		var bruteList []string
		for i := 1; i <= 50; i++ {
			bruteList = append(bruteList, strconv.Itoa(i))
		}
		return bruteList, nil
	}

	return ids, nil
}

func (s *SuperlogicaAdapter) checkUnit(ctx context.Context, id string, cpf string) (bool, *domain.Lead, error) {
	startTime := time.Now()
	
	endpoint := fmt.Sprintf("%s/unidades/index", s.apiURL)
	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)

	q := req.URL.Query()
	q.Add("idCondominio", id)
	q.Add("pesquisa", cpf)
	q.Add("itensPorPagina", "1")
	q.Add("exibirDadosDosContatos", "1")
	req.URL.RawQuery = q.Encode()

	s.addHeaders(req)

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

		return true, &domain.Lead{
			CPF:                   cpf,
			CondoID:               id,
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
