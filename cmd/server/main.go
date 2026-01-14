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
	"github.com/viplounge/platform/internal/handler"
	"github.com/viplounge/platform/internal/repository"
	"github.com/viplounge/platform/internal/service"
)

func main() {
	ctx := context.Background()

	// 1. Configuração
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")

	// 2. Dependências
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
	// Nota: Em caso de falha no repo, o service pode receber nil? 
	// Melhor criar um mock repo in-memory se repo for nil para dev local.
	// Para simplicidade do MVP, vamos passar repo mesmo que nil e deixar panicar ou tratar no service se chamado.
	// O ideal seria um FailoverRepository.
	svc := service.NewValidationService(repo, benefAdapter, partnerAdapter)

	// Handler
	h := handler.NewHandler(svc)
	
	// 3. Roteamento API
	// O método Routes retorna http.Handler, precisamos asserir se quisermos usar métodos específicos do Chi
	// ou apenas montar o handler api e montar o file server separadamente.
	// Vamos simplificar: Instanciar o router aqui.
	
	r := chi.NewRouter()
	
	// Mount API routes
	r.Mount("/", h.Routes())

	// 4. Servir Frontend Estático
	// Assumindo que o binário roda na raiz do projeto ou web/ está adjacente
	filesDir := "web" 
	if _, err := os.Stat(filesDir); os.IsNotExist(err) {
		// Tenta caminho relativo se rodando de cmd/server
		filesDir = "../../web"
	}
	
	// Serve static files
	fileServer(r, "/", http.Dir(filesDir))

	// 5. Iniciar Servidor
	log.Printf("Server starting on port %s", port)
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


