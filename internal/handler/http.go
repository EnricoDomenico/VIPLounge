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

	// Middlewares Básicos
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(customMiddleware.TenantMiddleware)  // Aplicar ANTES de tudo
	r.Use(customMiddleware.SecurityHeaders)
	
	// Configuração de CORS - incluir domínios específicos do VipLounge
	allowedOrigins := []string{
		"https://viplounge.com.br",
		"https://www.viplounge.com.br",
		"https://mobile.viplounge.com.br",
		"http://localhost:8080",
		"http://localhost:3000",
	}
	
	// Mesclar com origens da configuração, se houver
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

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Novo endpoint: GET /config - retorna configuração para frontend
	r.Get("/config", h.handleConfig)

	// Rotas da API - será chamada via Cloud Function proxy de Firebase
	r.Post("/v1/validate", h.handleValidate)
	r.Get("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Servir arquivos estáticos - images
	fs := http.FileServer(http.Dir("web"))
	r.Handle("/images/*", http.StripPrefix("/", fs))
	
	// Servir api-config.js com MIME type correto
	r.Get("/api-config.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "web/api-config.js")
	})
	
	// Servir index.html como raiz
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "web/index.html")
	})

	return r
}

var cpfRegex = regexp.MustCompile(`^\d{3}\.?\d{3}\.?\d{3}-?\d{2}$`)

func (h *Handler) handleValidate(w http.ResponseWriter, r *http.Request) {
	var req domain.ValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validação básica de formato
	if !cpfRegex.MatchString(req.CPF) {
		http.Error(w, "Invalid CPF format", http.StatusBadRequest)
		return
	}
	
	// IMPORTANTE: Buscar tenant_id do contexto (injetado pelo TenantMiddleware)
	// Isso garante que cada domínio use o ID de condomínio correto
	tenantID := customMiddleware.GetTenantID(r.Context())
	if tenantID != "" {
		req.CondoID = tenantID
	}
	
	// Fallback: Se não vier do contexto, usar configuração
	if req.CondoID == "" {
		req.CondoID = h.cfg.Behavior.DefaultCondoID
	}
	
	if h.cfg.Behavior.CondoIDRequired && req.CondoID == "" {
		http.Error(w, "Condo ID is required", http.StatusBadRequest)
		return
	}

	resp, err := h.svc.ValidateAndSave(r.Context(), req)
	if err != nil {
		// Log interno aqui
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}


