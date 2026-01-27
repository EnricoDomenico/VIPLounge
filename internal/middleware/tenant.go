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
	
	// Mapeamento de domínios específicos (opcional - para performance)
	// Se você souber antecipadamente o condomínio, pode mapear aqui
	switch host {
	case "viplounge.com.br", "www.viplounge.com.br":
		return "4" // Condomínio VIP Lounge específico (Unidade 48384 - Enrico)
	case "viplounge.mobile.adm.br":
		// BUSCA GLOBAL para mobile - permite encontrar qualquer morador
		log.Printf("[TENANT] Host mobile -> Usando busca global (tenant_id=-1)", host)
		return "-1"
	default:
		// BUSCA GLOBAL: Para hosts desconhecidos ou localhost, usar -1
		// Isso permite que a API Superlógica procure em todos os condomínios
		log.Printf("[TENANT] Host: %s -> Usando busca global (tenant_id=-1)", host)
		return "-1"
	}
}

// GetTenantID extrai o tenant_id do contexto
func GetTenantID(ctx context.Context) string {
	if tenantID, ok := ctx.Value(TenantIDKey).(string); ok {
		return tenantID
	}
	// Fallback: busca global se não houver tenant_id no contexto
	return "-1"
}
