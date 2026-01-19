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
		AppName        string `yaml:"app_name"`
		AppSubtitle    string `yaml:"app_subtitle"`
		AppLogo        string `yaml:"app_logo"`
		LogoURL        string `yaml:"logo_url"`
		SideImageURL   string `yaml:"side_image_url"`
		CompanyName    string `yaml:"company_name"`
		CompanyEmail   string `yaml:"company_email"`
		CompanyPhone   string `yaml:"company_phone"`
		ThemeColor     string `yaml:"theme_color"` // hex color
		SecondaryColor string `yaml:"secondary_color"`
	} `yaml:"branding"`

	// Mensagens customizáveis
	Messages struct {
		WelcomeMain       string `yaml:"welcome_main"`
		WelcomeHigh       string `yaml:"welcome_high"`
		WelcomeSubtext    string `yaml:"welcome_subtext"`
		CPFLabel          string `yaml:"cpf_label"`
		CPFPlaceholder    string `yaml:"cpf_placeholder"`
		FormTitle         string `yaml:"form_title"`
		SubmitButtonText  string `yaml:"submit_button_text"`
		ForgotPassword    string `yaml:"forgot_password"`
		NoAccount         string `yaml:"no_account"`
		SignupLink        string `yaml:"signup_link"`
		SuccessTitle      string `yaml:"success_title"`
		SuccessMessage    string `yaml:"success_message"`
		SuccessSubtext    string `yaml:"success_subtext"`
		AlreadyRegistered string `yaml:"already_registered"`
		NotFound          string `yaml:"not_found"`
		ErrorMessage      string `yaml:"error_message"`
		NetworkError      string `yaml:"network_error"`
		FooterText        string `yaml:"footer_text"`
	} `yaml:"messages"`

	// Comportamento
	Behavior struct {
		EnableDebugPanel      bool   `yaml:"enable_debug_panel"`
		Language              string `yaml:"language"`
		CondoIDRequired       bool   `yaml:"condo_id_required"`
		DefaultCondoID        string `yaml:"default_condo_id"`
		RedirectURLOnSuccess  string `yaml:"redirect_url_on_success"`
		RedirectURLOnError    string `yaml:"redirect_url_on_error"`
		ShowUserIDInModal     bool   `yaml:"show_user_id_in_modal"`
		ShowSideImage         bool   `yaml:"show_side_image"`
		AutoCloseModalSeconds int    `yaml:"auto_close_modal_seconds"`
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
	cfg.Branding.AppName = getEnvOrDefault("APP_NAME", "mobile")
	cfg.Branding.AppSubtitle = getEnvOrDefault("APP_SUBTITLE", "Acesso Exclusivo")
	cfg.Branding.LogoURL = getEnvOrDefault("LOGO_URL", "/images/logo.png")
	cfg.Branding.SideImageURL = getEnvOrDefault("SIDE_IMAGE_URL", "/images/ParteCondominio.png")
	cfg.Branding.CompanyName = getEnvOrDefault("COMPANY_NAME", "mobile")
	cfg.Branding.CompanyEmail = getEnvOrDefault("COMPANY_EMAIL", "contato@mobile.com")
	cfg.Branding.CompanyPhone = getEnvOrDefault("COMPANY_PHONE", "+55 11 9999-9999")
	cfg.Branding.ThemeColor = getEnvOrDefault("THEME_COLOR", "0066cc")
	cfg.Branding.SecondaryColor = getEnvOrDefault("SECONDARY_COLOR", "0052a3")

	// Messages (Portuguese as default)
	cfg.Messages.WelcomeMain = getEnvOrDefault("MSG_WELCOME_MAIN", "Bem-vindo ao seu")
	cfg.Messages.WelcomeHigh = getEnvOrDefault("MSG_WELCOME_HIGH", "espaço exclusivo")
	cfg.Messages.WelcomeSubtext = getEnvOrDefault("MSG_WELCOME_SUBTEXT", "Entre com suas credenciais para acessar")
	cfg.Messages.CPFLabel = getEnvOrDefault("MSG_CPF_LABEL", "CPF")
	cfg.Messages.CPFPlaceholder = getEnvOrDefault("MSG_CPF_PLACEHOLDER", "123.456.789-00")
	cfg.Messages.FormTitle = getEnvOrDefault("MSG_FORM_TITLE", "Área do Condômino")
	cfg.Messages.SubmitButtonText = getEnvOrDefault("MSG_SUBMIT_BTN", "Entrar")
	cfg.Messages.ForgotPassword = getEnvOrDefault("MSG_FORGOT_PASSWORD", "Esqueceu sua senha?")
	cfg.Messages.NoAccount = getEnvOrDefault("MSG_NO_ACCOUNT", "Não tem conta?")
	cfg.Messages.SignupLink = getEnvOrDefault("MSG_SIGNUP_LINK", "Cadastre-se agora")
	cfg.Messages.SuccessTitle = getEnvOrDefault("MSG_SUCCESS_TITLE", "PARABÉNS!")
	cfg.Messages.SuccessMessage = getEnvOrDefault("MSG_SUCCESS_MSG", "Você está participando!")
	cfg.Messages.SuccessSubtext = getEnvOrDefault("MSG_SUCCESS_SUBTEXT", "Aguarde o sorteio")
	cfg.Messages.AlreadyRegistered = getEnvOrDefault("MSG_ALREADY_REGISTERED", "Você já está cadastrado!")
	cfg.Messages.NotFound = getEnvOrDefault("MSG_NOT_FOUND", "CPF não encontrado.")
	cfg.Messages.ErrorMessage = getEnvOrDefault("MSG_ERROR", "Erro ao validar.")
	cfg.Messages.NetworkError = getEnvOrDefault("MSG_NETWORK_ERROR", "Erro de conexão com o servidor.")
	cfg.Messages.FooterText = getEnvOrDefault("MSG_FOOTER", "Plataforma segura e certificada")

	// Behavior
	cfg.Behavior.EnableDebugPanel = getEnvOrDefaultBool("ENABLE_DEBUG", false)
	cfg.Behavior.Language = getEnvOrDefault("LANGUAGE", "pt-BR")
	cfg.Behavior.CondoIDRequired = getEnvOrDefaultBool("CONDO_ID_REQUIRED", false)
	cfg.Behavior.DefaultCondoID = getEnvOrDefault("DEFAULT_CONDO_ID", "")
	cfg.Behavior.RedirectURLOnSuccess = getEnvOrDefault("REDIRECT_ON_SUCCESS", "")
	cfg.Behavior.RedirectURLOnError = getEnvOrDefault("REDIRECT_ON_ERROR", "")
	cfg.Behavior.ShowUserIDInModal = getEnvOrDefaultBool("SHOW_USER_ID", false)
	cfg.Behavior.ShowSideImage = getEnvOrDefaultBool("SHOW_SIDE_IMAGE", true)
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
