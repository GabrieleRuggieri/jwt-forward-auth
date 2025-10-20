package handler

import (
	"jwt-forward-auth/internal/auth"
	"log/slog"
	"net/http"
	"strings"
)

type AuthHandler struct {
	validator *auth.Validator
}

func NewAuthHandler(v *auth.Validator) *AuthHandler {
	return &AuthHandler{validator: v}
}

func (h *AuthHandler) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		slog.Warn("Header Authorization mancante")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		slog.Warn("Header Authorization malformato")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	tokenString := parts[1]

	claims, err := h.validator.ValidateToken(tokenString)
	if err != nil {
		// Il validator logga già l'errore specifico
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// **Funzionalità Chiave**: Inoltriamo l'identità dell'utente (dal claim 'sub')
	// al servizio a valle. Questo è estremamente utile.
	w.Header().Set("X-User-ID", claims.Subject)

	w.WriteHeader(http.StatusOK)
}