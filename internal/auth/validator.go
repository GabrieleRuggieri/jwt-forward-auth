package auth

import (
	"fmt"
	"jwt-forward-auth/internal/jwks"
	"log/slog"

	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims definisce i claims che ci aspettiamo.
type CustomClaims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

type Validator struct {
	jwksCache        *jwks.Cache
	allowedIssuers   []string
	allowedAudiences []string
	allowedAlg       string
}

func NewValidator(cache *jwks.Cache, issuers, audiences []string, alg string) *Validator {
	return &Validator{
		jwksCache:        cache,
		allowedIssuers:   issuers,
		allowedAudiences: audiences,
		allowedAlg:       alg,
	}
}

func (v *Validator) ValidateToken(tokenString string) (*CustomClaims, error) {
	// Il Keyfunc viene chiamato dalla libreria JWT durante il parsing.
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// 1. Controlla che l'algoritmo del token sia quello atteso.
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok && v.allowedAlg == "RS256" {
			// Aggiungi altri controlli per ES256, etc. se necessario
			return nil, fmt.Errorf("algoritmo di firma inaspettato: %v, atteso %v", token.Header["alg"], v.allowedAlg)
		}

		// 2. Estrai il Key ID ('kid') dall'header del token.
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("header 'kid' mancante nel token")
		}

		// 3. Ottieni la chiave pubblica dalla cache JWKS.
		key, err := v.jwksCache.GetKey(kid)
		if err != nil {
			return nil, fmt.Errorf("impossibile ottenere la chiave di validazione: %w", err)
		}
		return key, nil
	}

	// Opzioni di validazione per issuer e audience.
	// La libreria si occuperà di verificare che i claim nel token corrispondano.
	opts := []jwt.ParserOption{
		jwt.WithIssuer(v.allowedIssuers...),
		jwt.WithAudience(v.allowedAudiences...),
		jwt.WithValidMethods([]string{v.allowedAlg}),
	}

	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, keyFunc, opts...)

	if err != nil {
		slog.Warn("Validazione del token fallita", "error", err)
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		slog.Info("Token valido e verificato", "subject", claims.Subject, "issuer", claims.Issuer)
		return claims, nil
	}

	return nil, fmt.Errorf("token non valido")
}