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
	"sync"
	"time"

	"github.com/viplounge/platform/internal/domain"
)

// RedeParceriasClient integra com a API de Clube de Benefícios
type RedeParceriasClient struct {
	baseURL      string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	
	// Cache do token
	tokenMu    sync.RWMutex
	token      string
	tokenExp   time.Time
}

func NewClient() *RedeParceriasClient {
	url := os.Getenv("REDE_PARCERIAS_URL")
	if url == "" {
		url = "https://api.staging.clubeparcerias.com.br/api-client/v1"
	}

	clientID := os.Getenv("REDE_PARCERIAS_CLIENT_ID")
	if clientID == "" {
		clientID = "a08bb621-9bfc-46e4-ae5d-ea7e76610a9c"
	}

	clientSecret := os.Getenv("REDE_PARCERIAS_CLIENT_SECRET")
	if clientSecret == "" {
		clientSecret = "JKKKXFjVYr1MEW87LR6PmfehNUrwCrghZxY39Ja9"
	}

	return &RedeParceriasClient{
		baseURL:      strings.TrimRight(url, "/"),
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
	}
}

// getToken obtém token OAuth2 (client_credentials) com cache
func (c *RedeParceriasClient) getToken(ctx context.Context) (string, error) {
	// Verificar cache
	c.tokenMu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		token := c.token
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	// Gerar novo token via /auth
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
		"scope":         "*",
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/auth", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("erro criando request auth: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na chamada auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("auth falhou: HTTP %d - %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("erro decodificando auth response: %w", err)
	}

	// Guardar no cache (com margem de segurança de 5 min)
	c.tokenMu.Lock()
	c.token = result.AccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	c.tokenMu.Unlock()

	return result.AccessToken, nil
}

// FindUserByCPF verifica se um usuário existe na Rede Parcerias
func (c *RedeParceriasClient) FindUserByCPF(ctx context.Context, cpf string) (*domain.PartnerUser, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro obtendo token: %w", err)
	}

	// Limpar CPF
	cpfClean := regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")

	// Buscar usuário
	url := fmt.Sprintf("%s/users?search=%s&limit=1", c.baseURL, cpfClean)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro buscando usuário: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erro HTTP %d ao buscar usuário", resp.StatusCode)
	}

	var result struct {
		Data []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			CPF       string `json:"cpf"`
			Cellphone string `json:"cellphone"`
			Active    bool   `json:"active"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro decodificando response: %w", err)
	}

	// Verificar se encontrou o CPF exato
	for _, user := range result.Data {
		userCPF := regexp.MustCompile(`\D`).ReplaceAllString(user.CPF, "")
		if userCPF == cpfClean {
			return &domain.PartnerUser{
				ID:        user.ID,
				Name:      user.Name,
				Email:     user.Email,
				CPF:       user.CPF,
				Cellphone: user.Cellphone,
				Active:    user.Active,
			}, nil
		}
	}

	return nil, nil // Não encontrado
}

// RegisterUser cadastra um novo usuário na Rede Parcerias
func (c *RedeParceriasClient) RegisterUser(ctx context.Context, lead *domain.Lead) error {
	token, err := c.getToken(ctx)
	if err != nil {
		lead.RedeParceriasError = fmt.Sprintf("AUTH_ERROR: %v", err)
		lead.RedeParceriasStatus = domain.PartnerStatusFailed
		return fmt.Errorf("erro obtendo token: %w", err)
	}

	startTime := time.Now()

	// Limpar CPF
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

	if lead.Phone != "" {
		payload["cellphone"] = lead.Phone
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/users", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		lead.RedeParceriasError = fmt.Sprintf("REQUEST_ERROR: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		lead.RedeParceriasError = fmt.Sprintf("NETWORK_ERROR: %v", err)
		lead.RedeParceriasStatus = domain.PartnerStatusFailed
		return fmt.Errorf("erro na chamada: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	lead.RedeParceriasResponseMs = time.Since(startTime).Milliseconds()
	lead.RedeParceriasAttempts++

	// Sucesso
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Tentar extrair ID da resposta
		var respData map[string]interface{}
		if err := json.Unmarshal(respBody, &respData); err == nil {
			if id, ok := respData["id"].(string); ok {
				lead.RedeParceriasUserID = id
			}
		}
		lead.RedeParceriasStatus = domain.PartnerStatusRegistered
		lead.RedeParceriasError = ""
		return nil
	}

	// Usuário já existe (422)
	if resp.StatusCode == 422 {
		lead.RedeParceriasStatus = domain.PartnerStatusRegistered
		lead.RedeParceriasError = "USER_ALREADY_EXISTS"
		return nil
	}

	// Erro
	lead.RedeParceriasStatus = domain.PartnerStatusFailed
	lead.RedeParceriasError = fmt.Sprintf("HTTP_%d: %s", resp.StatusCode, string(respBody))
	return fmt.Errorf("erro cadastrando: HTTP %d", resp.StatusCode)
}

// DeleteUser desativa/remove um usuário da Rede Parcerias
func (c *RedeParceriasClient) DeleteUser(ctx context.Context, userID string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return fmt.Errorf("erro obtendo token: %w", err)
	}

	url := fmt.Sprintf("%s/users/%s", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("erro deletando usuário: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GetSSOToken gera um token SSO para login automático no clube
func (c *RedeParceriasClient) GetSSOToken(ctx context.Context, userID string) (*domain.SSOToken, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro obtendo token: %w", err)
	}

	url := fmt.Sprintf("%s/sso-token?user_id=%s", c.baseURL, userID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("erro gerando SSO token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result domain.SSOToken
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("erro decodificando SSO response: %w", err)
	}

	return &result, nil
}
