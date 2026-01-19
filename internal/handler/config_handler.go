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
		AppSubtitle    string `json:"app_subtitle"`
		LogoURL        string `json:"logo_url"`
		SideImageURL   string `json:"side_image_url"`
		CompanyName    string `json:"company_name"`
		CompanyEmail   string `json:"company_email"`
		CompanyPhone   string `json:"company_phone"`
		ThemeColor     string `json:"theme_color"`
		SecondaryColor string `json:"secondary_color"`
	} `json:"branding"`

	Messages struct {
		WelcomeMain       string `json:"welcome_main"`
		WelcomeHigh       string `json:"welcome_high"`
		WelcomeSubtext    string `json:"welcome_subtext"`
		CPFLabel          string `json:"cpf_label"`
		CPFPlaceholder    string `json:"cpf_placeholder"`
		FormTitle         string `json:"form_title"`
		SubmitButtonText  string `json:"submit_button_text"`
		ForgotPassword    string `json:"forgot_password"`
		NoAccount         string `json:"no_account"`
		SignupLink        string `json:"signup_link"`
		SuccessTitle      string `json:"success_title"`
		SuccessMessage    string `json:"success_message"`
		SuccessSubtext    string `json:"success_subtext"`
		AlreadyRegistered string `json:"already_registered"`
		NotFound          string `json:"not_found"`
		ErrorMessage      string `json:"error_message"`
		NetworkError      string `json:"network_error"`
		FooterText        string `json:"footer_text"`
	} `json:"messages"`

	Behavior struct {
		EnableDebugPanel     bool `json:"enable_debug_panel"`
		Language             string `json:"language"`
		CondoIDRequired      bool `json:"condo_id_required"`
		DefaultCondoID       string `json:"default_condo_id"`
		ShowUserIDInModal    bool `json:"show_user_id_in_modal"`
		ShowSideImage        bool `json:"show_side_image"`
		AutoCloseModalSeconds int `json:"auto_close_modal_seconds"`
	} `json:"behavior"`
}

// handleConfig retorna a configuração para o frontend
func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()

	resp := ConfigResponse{
		Branding: struct {
			AppName        string `json:"app_name"`
			AppSubtitle    string `json:"app_subtitle"`
			LogoURL        string `json:"logo_url"`
			SideImageURL   string `json:"side_image_url"`
			CompanyName    string `json:"company_name"`
			CompanyEmail   string `json:"company_email"`
			CompanyPhone   string `json:"company_phone"`
			ThemeColor     string `json:"theme_color"`
			SecondaryColor string `json:"secondary_color"`
		}{
			AppName:        cfg.Branding.AppName,
			AppSubtitle:    cfg.Branding.AppSubtitle,
			LogoURL:        cfg.Branding.LogoURL,
			SideImageURL:   cfg.Branding.SideImageURL,
			CompanyName:    cfg.Branding.CompanyName,
			CompanyEmail:   cfg.Branding.CompanyEmail,
			CompanyPhone:   cfg.Branding.CompanyPhone,
			ThemeColor:     cfg.Branding.ThemeColor,
			SecondaryColor: cfg.Branding.SecondaryColor,
		},
		Messages: struct {
			WelcomeMain       string `json:"welcome_main"`
			WelcomeHigh       string `json:"welcome_high"`
			WelcomeSubtext    string `json:"welcome_subtext"`
			CPFLabel          string `json:"cpf_label"`
			CPFPlaceholder    string `json:"cpf_placeholder"`
			FormTitle         string `json:"form_title"`
			SubmitButtonText  string `json:"submit_button_text"`
			ForgotPassword    string `json:"forgot_password"`
			NoAccount         string `json:"no_account"`
			SignupLink        string `json:"signup_link"`
			SuccessTitle      string `json:"success_title"`
			SuccessMessage    string `json:"success_message"`
			SuccessSubtext    string `json:"success_subtext"`
			AlreadyRegistered string `json:"already_registered"`
			NotFound          string `json:"not_found"`
			ErrorMessage      string `json:"error_message"`
			NetworkError      string `json:"network_error"`
			FooterText        string `json:"footer_text"`
		}{
			WelcomeMain:       cfg.Messages.WelcomeMain,
			WelcomeHigh:       cfg.Messages.WelcomeHigh,
			WelcomeSubtext:    cfg.Messages.WelcomeSubtext,
			CPFLabel:          cfg.Messages.CPFLabel,
			CPFPlaceholder:    cfg.Messages.CPFPlaceholder,
			FormTitle:         cfg.Messages.FormTitle,
			SubmitButtonText:  cfg.Messages.SubmitButtonText,
			ForgotPassword:    cfg.Messages.ForgotPassword,
			NoAccount:         cfg.Messages.NoAccount,
			SignupLink:        cfg.Messages.SignupLink,
			SuccessTitle:      cfg.Messages.SuccessTitle,
			SuccessMessage:    cfg.Messages.SuccessMessage,
			SuccessSubtext:    cfg.Messages.SuccessSubtext,
			AlreadyRegistered: cfg.Messages.AlreadyRegistered,
			NotFound:          cfg.Messages.NotFound,
			ErrorMessage:      cfg.Messages.ErrorMessage,
			NetworkError:      cfg.Messages.NetworkError,
			FooterText:        cfg.Messages.FooterText,
		},
		Behavior: struct {
			EnableDebugPanel     bool `json:"enable_debug_panel"`
			Language             string `json:"language"`
			CondoIDRequired      bool `json:"condo_id_required"`
			DefaultCondoID       string `json:"default_condo_id"`
			ShowUserIDInModal    bool `json:"show_user_id_in_modal"`
			ShowSideImage        bool `json:"show_side_image"`
			AutoCloseModalSeconds int `json:"auto_close_modal_seconds"`
		}{
			EnableDebugPanel:      cfg.Behavior.EnableDebugPanel,
			Language:              cfg.Behavior.Language,
			CondoIDRequired:       cfg.Behavior.CondoIDRequired,
			DefaultCondoID:        cfg.Behavior.DefaultCondoID,
			ShowUserIDInModal:     cfg.Behavior.ShowUserIDInModal,
			ShowSideImage:         cfg.Behavior.ShowSideImage,
			AutoCloseModalSeconds: cfg.Behavior.AutoCloseModalSeconds,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
