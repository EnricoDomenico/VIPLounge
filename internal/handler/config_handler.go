package handler

import (
	"encoding/json"
	"net/http"

	"github.com/viplounge/platform/internal/config"
)

// ConfigResponse estrutura de resposta do config endpoint
type ConfigResponse struct {
	Branding struct {
		AppName        string `json:"app_name"`
		AppLogo        string `json:"app_logo"`
		CompanyName    string `json:"company_name"`
		CompanyEmail   string `json:"company_email"`
		CompanyPhone   string `json:"company_phone"`
		ThemeColor     string `json:"theme_color"`
		SecondaryColor string `json:"secondary_color"`
	} `json:"branding"`

	Messages struct {
		WelcomeTitle      string `json:"welcome_title"`
		WelcomeSubtitle   string `json:"welcome_subtitle"`
		CPFLabel          string `json:"cpf_label"`
		CPFPlaceholder    string `json:"cpf_placeholder"`
		SubmitButtonText  string `json:"submit_button_text"`
		SuccessTitle      string `json:"success_title"`
		SuccessMessage    string `json:"success_message"`
		AlreadyRegistered string `json:"already_registered"`
		NotFound          string `json:"not_found"`
		ErrorMessage      string `json:"error_message"`
		NetworkError      string `json:"network_error"`
		SecurityDisclaim  string `json:"security_disclaimer"`
		FooterText        string `json:"footer_text"`
		ContactText       string `json:"contact_text"`
	} `json:"messages"`

	Behavior struct {
		EnableDebugPanel     bool   `json:"enable_debug_panel"`
		Language             string `json:"language"`
		DefaultCondoID       string `json:"default_condo_id"`
		ShowUserIDInModal    bool   `json:"show_user_id_in_modal"`
		AutoCloseModalSeconds int   `json:"auto_close_modal_seconds"`
	} `json:"behavior"`
}

// handleConfig retorna a configuração para o frontend
func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()

	resp := ConfigResponse{
		Branding: struct {
			AppName        string `json:"app_name"`
			AppLogo        string `json:"app_logo"`
			CompanyName    string `json:"company_name"`
			CompanyEmail   string `json:"company_email"`
			CompanyPhone   string `json:"company_phone"`
			ThemeColor     string `json:"theme_color"`
			SecondaryColor string `json:"secondary_color"`
		}{
			AppName:        cfg.Branding.AppName,
			AppLogo:        cfg.Branding.AppLogo,
			CompanyName:    cfg.Branding.CompanyName,
			CompanyEmail:   cfg.Branding.CompanyEmail,
			CompanyPhone:   cfg.Branding.CompanyPhone,
			ThemeColor:     cfg.Branding.ThemeColor,
			SecondaryColor: cfg.Branding.SecondaryColor,
		},
		Messages: struct {
			WelcomeTitle      string `json:"welcome_title"`
			WelcomeSubtitle   string `json:"welcome_subtitle"`
			CPFLabel          string `json:"cpf_label"`
			CPFPlaceholder    string `json:"cpf_placeholder"`
			SubmitButtonText  string `json:"submit_button_text"`
			SuccessTitle      string `json:"success_title"`
			SuccessMessage    string `json:"success_message"`
			AlreadyRegistered string `json:"already_registered"`
			NotFound          string `json:"not_found"`
			ErrorMessage      string `json:"error_message"`
			NetworkError      string `json:"network_error"`
			SecurityDisclaim  string `json:"security_disclaimer"`
			FooterText        string `json:"footer_text"`
			ContactText       string `json:"contact_text"`
		}{
			WelcomeTitle:      cfg.Messages.WelcomeTitle,
			WelcomeSubtitle:   cfg.Messages.WelcomeSubtitle,
			CPFLabel:          cfg.Messages.CPFLabel,
			CPFPlaceholder:    cfg.Messages.CPFPlaceholder,
			SubmitButtonText:  cfg.Messages.SubmitButtonText,
			SuccessTitle:      cfg.Messages.SuccessTitle,
			SuccessMessage:    cfg.Messages.SuccessMessage,
			AlreadyRegistered: cfg.Messages.AlreadyRegistered,
			NotFound:          cfg.Messages.NotFound,
			ErrorMessage:      cfg.Messages.ErrorMessage,
			NetworkError:      cfg.Messages.NetworkError,
			SecurityDisclaim:  cfg.Messages.SecurityDisclaimer,
			FooterText:        cfg.Messages.FooterText,
			ContactText:       cfg.Messages.ContactText,
		},
		Behavior: struct {
			EnableDebugPanel     bool   `json:"enable_debug_panel"`
			Language             string `json:"language"`
			DefaultCondoID       string `json:"default_condo_id"`
			ShowUserIDInModal    bool   `json:"show_user_id_in_modal"`
			AutoCloseModalSeconds int   `json:"auto_close_modal_seconds"`
		}{
			EnableDebugPanel:      cfg.Behavior.EnableDebugPanel,
			Language:              cfg.Behavior.Language,
			DefaultCondoID:        cfg.Behavior.DefaultCondoID,
			ShowUserIDInModal:     cfg.Behavior.ShowUserIDInModal,
			AutoCloseModalSeconds: cfg.Behavior.AutoCloseModalSeconds,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
