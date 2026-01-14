package handler

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/viplounge/platform/internal/domain"
	"github.com/viplounge/platform/internal/service"
)

type Handler struct {
	svc *service.ValidationService
}

func NewHandler(svc *service.ValidationService) *Handler {
	return &Handler{svc: svc}
}

// Routes define as rotas da aplicação
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	// Middlewares Básicos
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	// Configuração de CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // Em produção, restrinja ao domínio da Landing Page
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	r.Post("/v1/validate", h.handleValidate)

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
	if req.CondoID == "" {
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


