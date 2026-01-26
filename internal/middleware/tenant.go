package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
)

// ContextKey tipo para chaves de contexto
type ContextKey string

const (
	// TenantIDKey é a chave para armazenar o tenant_id no contexto
	TenantIDKey ContextKey = "tenant_id"
)

// TenantMiddleware captura o Host da requisição e injeta o tenant_id no contexto
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		
		// Remover porta se existir
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		
		// Mapeamento de domínios para tenant_id (ID do condomínio na Superlógica)
		tenantID := mapHostToTenantID(host)
		
		log.Printf("[TENANT] Host: %s -> Tenant ID: %s", host, tenantID)
		
		// Injetar tenant_id no contexto
		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// mapHostToTenantID mapeia o host para o ID do condomínio correspondente
func mapHostToTenantID(host string) string {
	// Normalizar host (lowercase)
	host = strings.ToLower(host)
	
	// Mapeamento de domínios
	switch host {
	case "viplounge.com.br", "www.viplounge.com.br":
		return "4" // Condomínio VIP Lounge (Unidade 48384 - Enrico)
	case "mobile.viplounge.com.br":
		return "4" // Mesmo condomínio
	case "localhost":
		return "4" // Desenvolvimento local
	default:
		// Fallback: se não reconhecer o domínio, usar ID 4 como padrão
		log.Printf("[TENANT] Host desconhecido: %s, usando tenant_id padrão: 4", host)
		return "4"
	}
}

// GetTenantID extrai o tenant_id do contexto
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	// Fallback se não houver tenant_id no contexto
	return "4"
}
