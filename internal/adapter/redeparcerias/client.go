package redeparcerias

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/viplounge/platform/internal/domain"
)

// RedeParceriasClient integra com a API de Clube de Benefícios
// Usa Bearer Token JWT para autenticação
type RedeParceriasClient struct {
	baseURL      string
	bearerToken  string
	httpClient   *http.Client
}

func NewClient() *RedeParceriasClient {
	url := os.Getenv("REDE_PARCERIAS_URL")
	if url == "" {
		url = "https://api.staging.clubeparcerias.com.br/api-client/v1"
	}

	token := os.Getenv("REDE_PARCERIAS_BEARER_TOKEN")
	if token == "" {
		// Token padrão (mudar em .env para produção)
		token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJhdWQiOiJhMDhiYjYyMS05YmZjLTQ2ZTQtYWU1ZC1lYTdlNzY2MTBhOWMiLCJqdGkiOiJmNDE3NDIzOWQzOTJkYTJlMzdmMzQzNTA1MjJiZWI5ODJiYzc0ZWY3NGM3OWE0MTkwMGZlODI2ODU0ZGI1NWMyNTdlYzk0ZWE5NGM4MWM5MCIsImlhdCI6MTc2ODI1MDk2Ni43NjU1MTYsIm5iZiI6MTc2ODI1MDk2Ni43NjU1MiwiZXhwIjoxNzk5Nzg2OTY2Ljc1NzM5NSwic3ViIjoiIiwic2NvcGVzIjpbIioiXX0.VWv2FN_LKsKB0YEJwrCtAO3BlOxylCKH5cX3gEuicw4kaK-UhZ419mHk6yXCLM0Sy7Qjvcn4Ps9bPCH3ndP1cA8WY0a_4qg2OnsX-6qPR_9HsNrm55S5lXHG1DGZ4hzKFl26dqo6E80WwdBVD3cL97P2eEoYYONhZqLdjdVPopcLVK6dvhL5sC2zXPtQ1VInKa8aVyHtQPxUN1G8fw1BkjHyQ-SBoV2Q8OAzX_HYI47qDsOPmgbaaAXLVI_VJRuTDhjEHmc5DbDByESx3NCxjWE0vUPaZNBWtoBYymjDJFm_Rc6todZGF_uPdD9XCOVx4Lj50HF4-XU8WZ29sfm1vEX0huuu0-1BnUiLLeizFJnO0K0UhGscf7yYFxp_QV7cKqrRP39efn980F3qJZbmDKg-_cC4Bogj7sFbxVURoY69ffpSFUksf61UEu_c4QHb60JwR_Z_M03YHDnR90GvZ0kvdpavOs85ADE0dNfeHvvHOEfJBDsJHfhyEimjL0n1Oj9BwD-xebw7OEwD3MbEmf5HKYqxjphwo0cUKZoIEMoWqOQkqmys5oiggNfTJ_lAbYBEPNkIX0zKMJxpCSMdhwUlp3gVt81CpfLDxjqxsJqypjJ0jmTIpvAPV0hkvHXxjcQGZYgG62ChqQqFmsYaF0TEFwEdoGRvqPb4RxsB6K4"
	}

	return &RedeParceriasClient{
		baseURL:     strings.TrimRight(url, "/"),
		bearerToken: token,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

// RegisterUser cadastra um usuário no clube de benefícios
// Implementa retry com exponential backoff
func (c *RedeParceriasClient) RegisterUser(ctx context.Context, lead *domain.Lead) error {
	// Configuração de retry
	maxAttempts := 3
	initialWait := time.Second
	maxWait := 30 * time.Second

	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		startTime := time.Now()

		// 1. Preparar payload
		cpfClean := regexp.MustCompile(`\D`).ReplaceAllString(lead.CPF, "")
		if len(cpfClean) > 11 {
			cpfClean = cpfClean[:11]
		}

		payload := map[string]interface{}{
			"name":       lead.Name,
			"email":      lead.Email,
			"cpf":        cpfClean,
			"authorized": true,
		}

		body, _ := json.Marshal(payload)

		// 2. Fazer requisição
		url := fmt.Sprintf("%s/users", c.baseURL)
		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
		if err != nil {
			lead.RedeParceriasError = fmt.Sprintf("REQUEST_BUILD_ERROR: %v", err)
			return err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.bearerToken)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			lead.RedeParceriasError = fmt.Sprintf("NETWORK_ERROR (attempt %d): %v", attempt, err)
			lead.RedeParceriasAttempts = attempt

			// Se não é última tentativa, aguardar e retry
			if attempt < maxAttempts {
				waitTime := exponentialBackoff(attempt, initialWait, maxWait)
				time.Sleep(waitTime)
				continue
			}
			lead.RedeParceriasStatus = domain.PartnerStatusFailed
			return fmt.Errorf("erro request rede parcerias (tentativas esgotadas): %w", err)
		}
		defer resp.Body.Close()

		// 3. Ler resposta completa (para logging)
		respBody, _ := io.ReadAll(resp.Body)
		responseTime := time.Since(startTime).Milliseconds()

		lead.RedeParceriasResponseMs = responseTime
		lead.RedeParceriasAttempts = attempt

		// 4. Processar resposta
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			// Sucesso! Extrair ID da resposta
			var respData map[string]interface{}
			if err := json.Unmarshal(respBody, &respData); err == nil {
				if userID, ok := respData["id"].(string); ok {
					lead.RedeParceriasUserID = userID
				}
			}
			lead.RedeParceriasStatus = domain.PartnerStatusRegistered
			lead.RedeParceriasError = ""
			return nil
		}

		if resp.StatusCode == 422 {
			// Já existe (tratado como sucesso)
			// Tentar extrair ID mesmo em 422
			var respData map[string]interface{}
			if err := json.Unmarshal(respBody, &respData); err == nil {
				if userID, ok := respData["id"].(string); ok {
					lead.RedeParceriasUserID = userID
				}
			}
			lead.RedeParceriasStatus = domain.PartnerStatusRegistered
			lead.RedeParceriasError = "USER_ALREADY_EXISTS (422)"
			return nil
		}

		if resp.StatusCode == 401 || resp.StatusCode == 403 {
			// Erro de autenticação - não vai mudar em retry
			lead.RedeParceriasStatus = domain.PartnerStatusFailed
			lead.RedeParceriasError = fmt.Sprintf("AUTH_ERROR: HTTP_%d", resp.StatusCode)
			return fmt.Errorf("erro autenticação rede parcerias: %d", resp.StatusCode)
		}

		// 5. Outros erros - pode tentar retry
		lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
		lead.RedeParceriasError = fmt.Sprintf("HTTP_%d (attempt %d)", resp.StatusCode, attempt)

		// Se for erro 5xx (servidor), pode tentar retry
		if resp.StatusCode >= 500 && attempt < maxAttempts {
			waitTime := exponentialBackoff(attempt, initialWait, maxWait)
			time.Sleep(waitTime)
			continue
		}

		// Erro 4xx que não é 422/401/403 - não vai mudar em retry
		lead.RedeParceriasStatus = domain.PartnerStatusFailed
		return fmt.Errorf("erro cadastro rede parcerias: %w", lastErr)
	}

	lead.RedeParceriasStatus = domain.PartnerStatusRetryPending
	return fmt.Errorf("rede parcerias: tentativas esgotadas - %w", lastErr)
}

// exponentialBackoff calcula o tempo de espera com backoff exponencial
// Exemplo: tentativa 1 = 1s, tentativa 2 = 2s, tentativa 3 = 4s
func exponentialBackoff(attempt int, initial, max time.Duration) time.Duration {
	waitTime := initial * time.Duration(1<<uint(attempt-1)) // 2^(attempt-1)
	if waitTime > max {
		waitTime = max
	}
	return waitTime
}

