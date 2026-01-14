package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config estrutura agnóstica de configuração
type Config struct {
	// Branding
	Branding struct {
		AppName      string `yaml:"app_name"`
		AppLogo      string `yaml:"app_logo"`
		CompanyName  string `yaml:"company_name"`
		CompanyEmail string `yaml:"company_email"`
		CompanyPhone string `yaml:"company_phone"`
		ThemeColor   string `yaml:"theme_color"` // hex color
		SecondaryColor string `yaml:"secondary_color"`
	} `yaml:"branding"`

	// Mensagens customizáveis
	Messages struct {
		WelcomeTitle       string `yaml:"welcome_title"`
		WelcomeSubtitle    string `yaml:"welcome_subtitle"`
		CPFLabel           string `yaml:"cpf_label"`
		CPFPlaceholder     string `yaml:"cpf_placeholder"`
		SubmitButtonText   string `yaml:"submit_button_text"`
		SuccessTitle       string `yaml:"success_title"`
		SuccessMessage     string `yaml:"success_message"`
		AlreadyRegistered  string `yaml:"already_registered"`
		NotFound           string `yaml:"not_found"`
		ErrorMessage       string `yaml:"error_message"`
		NetworkError       string `yaml:"network_error"`
		SecurityDisclaimer string `yaml:"security_disclaimer"`
		FooterText         string `yaml:"footer_text"`
		ContactText        string `yaml:"contact_text"`
	} `yaml:"messages"`

	// Comportamento
	Behavior struct {
		EnableDebugPanel      bool   `yaml:"enable_debug_panel"`
		Language              string `yaml:"language"` // pt-BR, en-US, es-ES
		CondoIDRequired       bool   `yaml:"condo_id_required"`
		DefaultCondoID        string `yaml:"default_condo_id"`
		RedirectURLOnSuccess  string `yaml:"redirect_url_on_success"`
		RedirectURLOnError    string `yaml:"redirect_url_on_error"`
		ShowUserIDInModal     bool   `yaml:"show_user_id_in_modal"`
		AutoCloseModalSeconds int    `yaml:"auto_close_modal_seconds"` // 0 = manual
	} `yaml:"behavior"`

	// Validação
	Validation struct {
		AllowedCPFFormats []string `yaml:"allowed_cpf_formats"` // "XXX.XXX.XXX-XX", "XXXXXXXXXXX"
		MaxRetries        int      `yaml:"max_retries"`
		RetryDelayMs      []int    `yaml:"retry_delay_ms"` // [1000, 2000, 4000]
	} `yaml:"validation"`

	// Segurança
	Security struct {
		CORSAllowedOrigins []string `yaml:"cors_allowed_origins"`
		RequireHTTPS       bool     `yaml:"require_https"`
		CSPEnabled         bool     `yaml:"csp_enabled"`
	} `yaml:"security"`

	// Integrações (agnósticas)
	Integrations struct {
		NameIntegration struct {
			Enabled bool   `yaml:"enabled"`
			Type    string `yaml:"type"` // "superlogica", "custom", etc
			URL     string `yaml:"url"`
		} `yaml:"name_integration"`
		PartnerIntegration struct {
			Enabled bool   `yaml:"enabled"`
			Type    string `yaml:"type"` // "rede_parcerias", "custom", etc
			URL     string `yaml:"url"`
		} `yaml:"partner_integration"`
	} `yaml:"integrations"`

	// Database
	Database struct {
		Type           string `yaml:"type"` // "firestore", "postgres", "mongodb"
		CollectionName string `yaml:"collection_name"`
	} `yaml:"database"`
}

var globalConfig *Config

// Load carrega a configuração de arquivo YAML e env vars
func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	// 1. Carregar defaults
	setDefaults(cfg)

	// 2. Carregar de arquivo YAML se existir
	if configPath != "" {
		if err := loadFromYAML(configPath, cfg); err != nil {
			return nil, fmt.Errorf("erro ao carregar config YAML: %w", err)
		}
	}

	// 3. Sobrescrever com variáveis de ambiente
	loadFromEnv(cfg)

	globalConfig = cfg
	return cfg, nil
}

// Get retorna a instância global de config
func Get() *Config {
	if globalConfig == nil {
		globalConfig = &Config{}
		setDefaults(globalConfig)
	}
	return globalConfig
}

func setDefaults(cfg *Config) {
	// Branding
	cfg.Branding.AppName = getEnvOrDefault("APP_NAME", "VIP Lounge")
	cfg.Branding.CompanyName = getEnvOrDefault("COMPANY_NAME", "VIP Lounge")
	cfg.Branding.CompanyEmail = getEnvOrDefault("COMPANY_EMAIL", "contato@viplounge.com")
	cfg.Branding.CompanyPhone = getEnvOrDefault("COMPANY_PHONE", "+55 11 9999-9999")
	cfg.Branding.ThemeColor = getEnvOrDefault("THEME_COLOR", "4f46e5") // indigo-600
	cfg.Branding.SecondaryColor = getEnvOrDefault("SECONDARY_COLOR", "8b5cf6") // purple-500

	// Messages (Portuguese as default)
	cfg.Messages.WelcomeTitle = getEnvOrDefault("MSG_WELCOME_TITLE", "Bem-vindo")
	cfg.Messages.WelcomeSubtitle = getEnvOrDefault("MSG_WELCOME_SUBTITLE", "Valide seu acesso exclusivo inserindo seu CPF abaixo.")
	cfg.Messages.CPFLabel = getEnvOrDefault("MSG_CPF_LABEL", "CPF do Titular")
	cfg.Messages.CPFPlaceholder = getEnvOrDefault("MSG_CPF_PLACEHOLDER", "000.000.000-00")
	cfg.Messages.SubmitButtonText = getEnvOrDefault("MSG_SUBMIT_BTN", "Validar Acesso")
	cfg.Messages.SuccessTitle = getEnvOrDefault("MSG_SUCCESS_TITLE", "PARABÉNS!")
	cfg.Messages.SuccessMessage = getEnvOrDefault("MSG_SUCCESS_MSG", "Bem-vindo ao Clube!")
	cfg.Messages.AlreadyRegistered = getEnvOrDefault("MSG_ALREADY_REGISTERED", "Você já está cadastrado em nosso clube de beneficiários!")
	cfg.Messages.NotFound = getEnvOrDefault("MSG_NOT_FOUND", "Condomínio não participante ou CPF não encontrado.")
	cfg.Messages.ErrorMessage = getEnvOrDefault("MSG_ERROR", "Erro ao cadastrar no clube. Tente novamente.")
	cfg.Messages.NetworkError = getEnvOrDefault("MSG_NETWORK_ERROR", "Erro ao conectar com o servidor. Verifique sua conexão.")
	cfg.Messages.SecurityDisclaimer = getEnvOrDefault("MSG_SECURITY", "Seus dados estão seguros e criptografados.")
	cfg.Messages.FooterText = getEnvOrDefault("MSG_FOOTER", "2026 VIP Lounge Platform. All rights reserved.")
	cfg.Messages.ContactText = getEnvOrDefault("MSG_CONTACT", "Fale Conosco")

	// Behavior
	cfg.Behavior.EnableDebugPanel = getEnvOrDefaultBool("ENABLE_DEBUG", true)
	cfg.Behavior.Language = getEnvOrDefault("LANGUAGE", "pt-BR")
	cfg.Behavior.CondoIDRequired = getEnvOrDefaultBool("CONDO_ID_REQUIRED", false)
	cfg.Behavior.DefaultCondoID = getEnvOrDefault("DEFAULT_CONDO_ID", "")
	cfg.Behavior.RedirectURLOnSuccess = getEnvOrDefault("REDIRECT_ON_SUCCESS", "")
	cfg.Behavior.RedirectURLOnError = getEnvOrDefault("REDIRECT_ON_ERROR", "")
	cfg.Behavior.ShowUserIDInModal = getEnvOrDefaultBool("SHOW_USER_ID", true)
	cfg.Behavior.AutoCloseModalSeconds = getEnvOrDefaultInt("AUTO_CLOSE_MODAL_SECONDS", 0)

	// Validation
	cfg.Validation.AllowedCPFFormats = []string{"XXX.XXX.XXX-XX", "XXXXXXXXXXX"}
	cfg.Validation.MaxRetries = getEnvOrDefaultInt("MAX_RETRIES", 3)
	cfg.Validation.RetryDelayMs = []int{1000, 2000, 4000}

	// Security
	cfg.Security.CORSAllowedOrigins = []string{"*"}
	if origins := os.Getenv("CORS_ORIGINS"); origins != "" {
		// Parse comma-separated origins
		cfg.Security.CORSAllowedOrigins = []string{origins}
	}
	cfg.Security.RequireHTTPS = getEnvOrDefaultBool("REQUIRE_HTTPS", false)
	cfg.Security.CSPEnabled = getEnvOrDefaultBool("CSP_ENABLED", false)

	// Integrations
	cfg.Integrations.NameIntegration.Enabled = true
	cfg.Integrations.NameIntegration.Type = "superlogica"
	cfg.Integrations.NameIntegration.URL = getEnvOrDefault("SUPERLOGICA_URL", "https://api.superlogica.net/v2/condor")

	cfg.Integrations.PartnerIntegration.Enabled = true
	cfg.Integrations.PartnerIntegration.Type = "rede_parcerias"
	cfg.Integrations.PartnerIntegration.URL = getEnvOrDefault("REDE_PARCERIAS_URL", "https://api.staging.clubeparcerias.com.br/api-client/v1")

	// Database
	cfg.Database.Type = getEnvOrDefault("DB_TYPE", "firestore")
	cfg.Database.CollectionName = getEnvOrDefault("DB_COLLECTION_NAME", "leads")
}

func loadFromYAML(filePath string, cfg *Config) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return err
	}

	return nil
}

func loadFromEnv(cfg *Config) {
	// Sobrescrever com env vars (menor prioridade que YAML)
	if val := os.Getenv("APP_NAME"); val != "" {
		cfg.Branding.AppName = val
	}
	if val := os.Getenv("COMPANY_NAME"); val != "" {
		cfg.Branding.CompanyName = val
	}
	if val := os.Getenv("THEME_COLOR"); val != "" {
		cfg.Branding.ThemeColor = val
	}
}

// Helper functions
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvOrDefaultBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		b, _ := strconv.ParseBool(val)
		return b
	}
	return defaultVal
}

func getEnvOrDefaultInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		i, _ := strconv.Atoi(val)
		return i
	}
	return defaultVal
}
