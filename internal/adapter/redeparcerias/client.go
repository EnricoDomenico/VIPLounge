package redeparcerias

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

	// Token fixo (se fornecido, pula OAuth2)
	fixedToken string

	// Cache do token (para OAuth2)
	tokenMu  sync.RWMutex
	token    string
	tokenExp time.Time
}

func NewClient() *RedeParceriasClient {
	baseURL := os.Getenv("REDE_PARCERIAS_URL")
	if baseURL == "" {
		// PRODUÇÃO
		baseURL = "https://infratech.clubeparcerias.com.br/api-client/v1"
	}

	// OPÇÃO 1: Token fixo via variável de ambiente (para casos especiais)
	fixedToken := os.Getenv("REDE_PARCERIAS_BEARER_TOKEN")

	// OPÇÃO 2 (RECOMENDADO): OAuth2 client_credentials - gera token automaticamente
	clientID := os.Getenv("REDE_PARCERIAS_CLIENT_ID")
	if clientID == "" {
		// PRODUÇÃO
		clientID = "a08bba23-53aa-46ad-9565-ce7e1fdb169c"
	}

	clientSecret := os.Getenv("REDE_PARCERIAS_CLIENT_SECRET")
	if clientSecret == "" {
		// PRODUÇÃO
		clientSecret = "D9pgDnw3VwqOtij9Y5aU181zXmwOQyuMEWSVZqwh"
	}

	authMode := "OAuth2"
	if fixedToken != "" {
		authMode = "Token Fixo"
	}
	log.Printf("[REDE_PARCERIAS] Inicializado - URL: %s, Auth: %s", baseURL, authMode)

	return &RedeParceriasClient{
		baseURL:      strings.TrimRight(baseURL, "/"),
		fixedToken:   fixedToken, // Vazio = usar OAuth2 automaticamente
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 20 * time.Second},
	}
}

// getToken obtém token para autenticação
func (c *RedeParceriasClient) getToken(ctx context.Context) (string, error) {
	// PRIORIDADE 1: Token fixo
	if c.fixedToken != "" {
		return c.fixedToken, nil
	}

	// PRIORIDADE 2: Cache
	c.tokenMu.RLock()
	if c.token != "" && time.Now().Before(c.tokenExp) {
		token := c.token
		c.tokenMu.RUnlock()
		return token, nil
	}
	c.tokenMu.RUnlock()

	// PRIORIDADE 3: OAuth2
	payload := map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
		"scope":         "*",
	}

	body, _ := json.Marshal(payload)
	reqURL := fmt.Sprintf("%s/auth", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(body))
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

	// Cache com margem de 5 min
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

	cpfClean := regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")
	reqURL := fmt.Sprintf("%s/users?search=%s&limit=5", c.baseURL, cpfClean)

	log.Printf("[REDE_PARCERIAS] Buscando CPF: %s", cpfClean)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[REDE_PARCERIAS] FindUser Response: %d - %s", resp.StatusCode, string(respBody))

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

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("erro decodificando response: %w", err)
	}

	// Verificar se encontrou o CPF exato
	for _, user := range result.Data {
		userCPF := regexp.MustCompile(`\D`).ReplaceAllString(user.CPF, "")
		if userCPF == cpfClean {
			log.Printf("[REDE_PARCERIAS] Usuário encontrado: ID=%s, Email=%s", user.ID, user.Email)
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

	log.Printf("[REDE_PARCERIAS] CPF %s não encontrado", cpfClean)
	return nil, nil
}

// RegisterUser cadastra um novo usuário na Rede Parcerias COM authorized:true
func (c *RedeParceriasClient) RegisterUser(ctx context.Context, lead *domain.Lead) error {
	token, err := c.getToken(ctx)
	if err != nil {
		lead.RedeParceriasError = fmt.Sprintf("AUTH_ERROR: %v", err)
		lead.RedeParceriasStatus = domain.PartnerStatusFailed
		return fmt.Errorf("erro obtendo token: %w", err)
	}

	startTime := time.Now()

	// CPF: apenas números, máximo 11 caracteres
	cpfClean := regexp.MustCompile(`\D`).ReplaceAllString(lead.CPF, "")
	if len(cpfClean) > 11 {
		cpfClean = cpfClean[:11]
	}

	// Payload conforme documentação da API
	payload := map[string]interface{}{
		"name":       strings.TrimSpace(lead.Name),
		"email":      strings.TrimSpace(lead.Email),
		"cpf":        cpfClean,
		"authorized": true, // IMPORTANTE: Ativa o usuário imediatamente
	}

	// Celular é opcional - formato: (XX) 9XXXX-XXXX ou apenas números com 11 dígitos
	if lead.Phone != "" {
		phoneClean := regexp.MustCompile(`\D`).ReplaceAllString(lead.Phone, "")
		// Validar: deve ter 11 dígitos e começar com DDD válido (11-99) + 9
		if len(phoneClean) == 11 && phoneClean[2] == '9' {
			// Formatar como (XX) 9XXXX-XXXX para a API aceitar
			formattedPhone := fmt.Sprintf("(%s) %s-%s", 
				phoneClean[0:2], 
				phoneClean[2:7], 
				phoneClean[7:11])
			payload["cellphone"] = formattedPhone
			log.Printf("[REDE_PARCERIAS] Celular formatado: %s", formattedPhone)
		}
	}

	body, _ := json.Marshal(payload)
	reqURL := fmt.Sprintf("%s/users", c.baseURL)

	log.Printf("[REDE_PARCERIAS] Cadastrando usuário: %s", string(body))

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(body))
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

	log.Printf("[REDE_PARCERIAS] RegisterUser Response: %d - %s", resp.StatusCode, string(respBody))

	// Sucesso: 200 ou 201
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var respData map[string]interface{}
		if err := json.Unmarshal(respBody, &respData); err == nil {
			if id, ok := respData["id"].(string); ok {
				lead.RedeParceriasUserID = id
				log.Printf("[REDE_PARCERIAS] Usuário criado com ID: %s", id)
			}
		}
		lead.RedeParceriasStatus = domain.PartnerStatusRegistered
		lead.RedeParceriasError = ""
		return nil
	}

	// 422 = Pode ser "usuário já existe" OU erro de validação
	if resp.StatusCode == 422 {
		var errorResp struct {
			Message string            `json:"message"`
			Errors  map[string][]string `json:"errors"`
		}
		json.Unmarshal(respBody, &errorResp)
		
		// Verificar se é realmente "usuário já existe"
		isUserExists := strings.Contains(strings.ToLower(errorResp.Message), "email") && 
			strings.Contains(strings.ToLower(errorResp.Message), "cadastrado") ||
			strings.Contains(strings.ToLower(errorResp.Message), "cpf") && 
			strings.Contains(strings.ToLower(errorResp.Message), "cadastrado") ||
			strings.Contains(strings.ToLower(errorResp.Message), "already") ||
			strings.Contains(strings.ToLower(errorResp.Message), "existe")
		
		if isUserExists {
			log.Printf("[REDE_PARCERIAS] Usuário já existe (422)")
			lead.RedeParceriasStatus = domain.PartnerStatusRegistered
			lead.RedeParceriasError = "USER_ALREADY_EXISTS"
			return nil
		}
		
		// É erro de validação - tentar novamente sem o campo problemático
		log.Printf("[REDE_PARCERIAS] Erro de validação 422: %s", errorResp.Message)
		lead.RedeParceriasStatus = domain.PartnerStatusFailed
		lead.RedeParceriasError = fmt.Sprintf("VALIDATION_ERROR: %s", errorResp.Message)
		return fmt.Errorf("erro de validação: %s", errorResp.Message)
	}

	// Erro
	lead.RedeParceriasStatus = domain.PartnerStatusFailed
	lead.RedeParceriasError = fmt.Sprintf("HTTP_%d: %s", resp.StatusCode, string(respBody))
	return fmt.Errorf("erro cadastrando: HTTP %d - %s", resp.StatusCode, string(respBody))
}

// GetSSOToken gera token SSO para login automático
// IMPORTANTE: user_id pode ser o UUID ou o EMAIL do usuário
func (c *RedeParceriasClient) GetSSOToken(ctx context.Context, userIdentifier string) (*domain.SSOToken, error) {
	token, err := c.getToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("erro obtendo token: %w", err)
	}

	// Encode o identificador (importante se for email com @)
	encodedID := url.QueryEscape(userIdentifier)
	reqURL := fmt.Sprintf("%s/sso-token?user_id=%s", c.baseURL, encodedID)

	log.Printf("[REDE_PARCERIAS] Gerando SSO para: %s", userIdentifier)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[REDE_PARCERIAS] SSO Response: %d - %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result domain.SSOToken
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("erro decodificando SSO response: %w", err)
	}

	log.Printf("[REDE_PARCERIAS] SSO Token gerado! Redirect: %s", result.Redirect)
	return &result, nil
}

// RegisterAndGetSSO é o método principal que faz o fluxo completo:
// 1. Cadastra o usuário (com authorized:true)
// 2. Gera o SSO Token usando o EMAIL
// 3. Retorna a URL de redirect para login automático
func (c *RedeParceriasClient) RegisterAndGetSSO(ctx context.Context, lead *domain.Lead) (*domain.SSOToken, error) {
	log.Printf("[REDE_PARCERIAS] === FLUXO COMPLETO: RegisterAndGetSSO ===")
	log.Printf("[REDE_PARCERIAS] Nome: %s, Email: %s, CPF: %s", lead.Name, lead.Email, lead.CPF)

	// PASSO 1: Cadastrar usuário
	if err := c.RegisterUser(ctx, lead); err != nil {
		// Se não for erro de "já existe", retorna o erro
		if lead.RedeParceriasError != "USER_ALREADY_EXISTS" {
			return nil, fmt.Errorf("erro no cadastro: %w", err)
		}
		log.Printf("[REDE_PARCERIAS] Usuário já existia, continuando para SSO...")
	}

	// PASSO 2: Gerar SSO Token usando o EMAIL (mais confiável que ID)
	// A API aceita tanto UUID quanto email como user_id
	ssoIdentifier := lead.Email
	if ssoIdentifier == "" {
		return nil, fmt.Errorf("email é obrigatório para gerar SSO")
	}

	sso, err := c.GetSSOToken(ctx, ssoIdentifier)
	if err != nil {
		log.Printf("[REDE_PARCERIAS] Erro ao gerar SSO com email, tentando com ID...")
		
		// Fallback: tentar com o ID se tiver
		if lead.RedeParceriasUserID != "" {
			sso, err = c.GetSSOToken(ctx, lead.RedeParceriasUserID)
			if err != nil {
				return nil, fmt.Errorf("erro gerando SSO: %w", err)
			}
		} else {
			return nil, fmt.Errorf("erro gerando SSO: %w", err)
		}
	}

	log.Printf("[REDE_PARCERIAS] === SUCESSO! Redirect URL: %s ===", sso.Redirect)
	return sso, nil
}

// DeleteUser desativa/remove um usuário da Rede Parcerias
func (c *RedeParceriasClient) DeleteUser(ctx context.Context, userID string) error {
	token, err := c.getToken(ctx)
	if err != nil {
		return fmt.Errorf("erro obtendo token: %w", err)
	}

	reqURL := fmt.Sprintf("%s/users/%s", c.baseURL, userID)

	log.Printf("[REDE_PARCERIAS] Deletando usuário: %s", userID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", reqURL, nil)
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

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[REDE_PARCERIAS] DeleteUser Response: %d - %s", resp.StatusCode, string(respBody))

	if resp.StatusCode != 200 && resp.StatusCode != 404 {
		return fmt.Errorf("erro HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
