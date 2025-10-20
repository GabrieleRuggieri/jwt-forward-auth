package jwks

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwk"
)

// Cache gestisce il set di chiavi JWKS in memoria.
type Cache struct {
	mu         sync.RWMutex
	keySet     jwk.Set
	jwksURL    string
	ttl        time.Duration
	lastFetch  time.Time
	fetchCtx   context.Context
}

// NewCache crea una nuova cache e esegue il primo fetch delle chiavi.
func NewCache(ctx context.Context, jwksURL string, ttl time.Duration) (*Cache, error) {
	c := &Cache{
		jwksURL: jwksURL,
		ttl:     ttl,
		fetchCtx: ctx,
	}

	if err := c.refreshKeys(); err != nil {
		return nil, fmt.Errorf("impossibile eseguire il fetch iniziale delle chiavi JWKS: %w", err)
	}
	slog.Info("Chiavi JWKS caricate e messe in cache con successo", "url", jwksURL)
	return c, nil
}

// GetKey cerca una chiave pubblica tramite il suo Key ID (kid).
// Se la cache è scaduta, tenta di aggiornarla prima di cercare la chiave.
func (c *Cache) GetKey(kid string) (interface{}, error) {
	c.mu.RLock()
	// Se la cache è ancora valida, procedi
	if time.Since(c.lastFetch) < c.ttl {
		key, err := c.findKey(kid)
		c.mu.RUnlock()
		return key, err
	}
	c.mu.RUnlock()

	// La cache è scaduta, è necessario un refresh
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check: un'altra goroutine potrebbe aver già aggiornato la cache
	// mentre si attendeva il lock.
	if time.Since(c.lastFetch) < c.ttl {
		return c.findKey(kid)
	}

	if err := c.refreshKeys(); err != nil {
		// In caso di errore, proviamo a usare le chiavi vecchie (stale) se disponibili
		slog.Warn("Impossibile aggiornare le chiavi JWKS, si tenta di usare la cache precedente", "error", err)
		if c.keySet != nil {
			return c.findKey(kid)
		}
		return nil, err
	}

	slog.Info("Cache delle chiavi JWKS aggiornata con successo")
	return c.findKey(kid)
}

// findKey cerca la chiave nel set attuale (deve essere chiamato all'interno di un lock).
func (c *Cache) findKey(kid string) (interface{}, error) {
	key, ok := c.keySet.LookupKeyID(kid)
	if !ok {
		return nil, fmt.Errorf("chiave con kid '%s' non trovata nel set JWKS", kid)
	}

	var rawKey interface{}
	if err := key.Raw(&rawKey); err != nil {
		return nil, fmt.Errorf("impossibile estrarre la chiave pubblica dal JWK: %w", err)
	}

	return rawKey, nil
}

// refreshKeys esegue la richiesta HTTP per aggiornare il set di chiavi.
func (c *Cache) refreshKeys() error {
	keySet, err := jwk.Fetch(c.fetchCtx, c.jwksURL)
	if err != nil {
		return fmt.Errorf("fallito il fetch dall'URL JWKS '%s': %w", c.jwksURL, err)
	}

	c.keySet = keySet
	c.lastFetch = time.Now()
	return nil
}