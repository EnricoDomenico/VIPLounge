package middleware

import (
	"net/http"
)

// SecurityHeaders adiciona headers HTTP de segurança
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// XSS Protection (deprecated mas ainda útil)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy (restritivo)
		w.Header().Set("Content-Security-Policy", 
			"default-src 'self'; script-src 'self' 'unsafe-inline' https://fonts.googleapis.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self'")

		// HSTS (HTTP Strict Transport Security)
		// Em produção, aumentar para 31536000 (1 ano)
		w.Header().Set("Strict-Transport-Security", "max-age=3600; includeSubDomains; preload")

		// Permissão de Feature Policy
		w.Header().Set("Permissions-Policy", 
			"geolocation=(), microphone=(), camera=(), payment=()")

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware pode ser adicionado depois para rate limiting
// Por enquanto, usando Chi's built-in throttle
