package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/viplounge/platform/internal/adapter/benef"
	"github.com/viplounge/platform/internal/adapter/redeparcerias"
	"github.com/viplounge/platform/internal/config"
	"github.com/viplounge/platform/internal/handler"
	"github.com/viplounge/platform/internal/repository"
	"github.com/viplounge/platform/internal/service"
)

func main() {
	ctx := context.Background()

	// 1. Carregar Configuração Agnóstica
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Printf("WARN: Erro carregando config.yaml, usando defaults: %v", err)
		cfg = config.Get()
	}
	log.Printf("App carregado: %s", cfg.Branding.AppName)

	// 2. Configuração
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	// 3. Dependências
	// Repo
	repo, err := repository.NewFirestoreRepository(ctx, projectID)
	if err != nil {
		log.Printf("WARN: Firestore init failed (expected in local dev without creds): %v", err)
	} else {
		defer repo.Close()
	}

	// Adapter
	benefAdapter := benef.NewBenefAdapter()
	partnerAdapter := redeparcerias.NewClient()

	// Service
	svc := service.NewValidationService(repo, benefAdapter, partnerAdapter, cfg)

	// Handler
	h := handler.NewHandler(svc, cfg)
	
	// 4. Roteamento API
	r := chi.NewRouter()
	
	// Mount API routes (Handler contém CORS)
	r.Mount("/", h.Routes())

	// 5. Servir Frontend Estático
	filesDir := "web" 
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		filesDir = "../../web"
	}
	
	// Serve static files
	fileServer(r, "/", http.Dir(filesDir))

	// 6. Iniciar Servidor
	log.Printf("🚀 Server '%s' starting on port %s", cfg.Branding.AppName, port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}


