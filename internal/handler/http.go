package handler

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/viplounge/platform/internal/config"
	"github.com/viplounge/platform/internal/domain"
	customMiddleware "github.com/viplounge/platform/internal/middleware"
	"github.com/viplounge/platform/internal/service"
)

type Handler struct {
	svc *service.ValidationService
	cfg *config.Config
}

func NewHandler(svc *service.ValidationService, cfg *config.Config) *Handler {
	return &Handler{svc: svc, cfg: cfg}
}

// Routes define as rotas da aplicação
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	// 1. HEARTBEAT - Otimização para o Cron-job
	// Deve ser o primeiro middleware para responder rápido e com corpo mínimo (.)
	r.Use(middleware.Heartbeat("/v1/health"))
	r.Use(middleware.Heartbeat("/health"))

	// 2. Middlewares de Base e Multi-tenancy
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.TenantMiddleware) // Identifica o condomínio pelo Host
	r.Use(customMiddleware.SecurityHeaders)
	
	// 3. Configuração de CORS para os domínios oficiais
	allowedOrigins := []string{
		"https://viplounge.com.br",
		"https://viplounge.mobile.adm.br",
		"https://www.viplounge.com.br",
		"https://mobile.viplounge.com.br",
		"http://localhost:8080",
		"http://localhost:3000",
	}
	
	if len(h.cfg.Security.CORSAllowedOrigins) > 0 && h.cfg.Security.CORSAllowedOrigins[0] != "*" {
		allowedOrigins = append(allowedOrigins, h.cfg.Security.CORSAllowedOrigins...)
	}
	
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// 4. Endpoints de Configuração e API
	r.Get("/config", h.handleConfig)
	r.Post("/v1/validate", h.handleValidate)
	r.Post("/v1/confirm-email", h.handleConfirmEmail)

	// 5. Servir Arquivos Estáticos (Substituindo o Firebase)
	// File Server para arquivos estáticos
	fs := http.FileServer(http.Dir("web"))
	
	// Rotas específicas para arquivos conhecidos (otimização + headers corretos)
	r.Get("/api-config.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		http.ServeFile(w, r, "web/api-config.js")
	})
	
	r.Get("/backend-config.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		http.ServeFile(w, r, "web/backend-config.json")
	})
	
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/favicon.ico")
	})
	
	// Rota Raiz: Serve o portal do cliente. Essencial para validação de SSL
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "web/index.html")
	})
	
	// CATCH-ALL: Servir todos os outros arquivos estáticos (imagens, CSS, etc)
	// Deve ser a ÚLTIMA rota para não conflitar com as rotas da API
	r.Handle("/*", http.StripPrefix("/", fs))

	return r
}

var cpfRegex = regexp.MustCompile(`^\d{3}\.?\d{3}\.?\d{3}-?\d{2}$`)

func (h *Handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	var req domain.ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !cpfRegex.MatchString(req.CPF) {
		http.Error(w, "Invalid CPF format", http.StatusBadRequest)
		return
	}
	
	// PRIORIDADE DE BUSCA:
	// 1. Se CondoID vier vazio no JSON, usar o Tenant ID do contexto (pode ser -1 ou específico)
	// 2. Se o contexto também estiver vazio, usar DefaultCondoID da config
	if req.CondoID == "" {
		tenantID := customMiddleware.GetTenantID(r.Context())
		if tenantID != "" {
			req.CondoID = tenantID
		} else {
			req.CondoID = h.cfg.Behavior.DefaultCondoID
		}
	}
	
	if h.cfg.Behavior.CondoIDRequired && req.CondoID == "" {
		http.Error(w, "Condo ID is required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.ValidateAndSave(r.Context(), req)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) handleConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var req domain.EmailConfirmationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !cpfRegex.MatchString(req.CPF) {
		http.Error(w, "Invalid CPF format", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.ConfirmEmailAndActivate(r.Context(), req)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}