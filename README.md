# Go JWT Forward Authentication Service for Traefik

Questo è un microservizio Go ad alte prestazioni per la validazione di token JWT, progettato per integrarsi con Traefik tramite il middleware `ForwardAuth`.

Questo servizio replica le funzionalità di validazione di base di un middleware JWT tradizionale (come `tpaulus/jwt-middleware`) ma con i vantaggi di un'architettura a microservizi.

## ✨ Caratteristiche

- **Validazione JWT basata su JWKS**: Scarica e mette in cache le chiavi pubbliche da un URL (es. Auth0, Okta, Keycloak) per una validazione sicura della firma.
- **Caching Intelligente delle Chiavi**: Le chiavi JWKS sono tenute in cache con un TTL configurabile per massimizzare le performance e la resilienza.
- **Validazione Rigorosa dei Claim**: Controlla `issuer`, `audience` e `algorithm` del token in base alla configurazione.
- **Inoltro dell'Identità Utente**: Passa in modo sicuro l'ID dell'utente (dal claim `sub`) al servizio a valle tramite l'header `X-User-ID`.
- **Architettura Forward Auth**: Si integra con Traefik in modo pulito e disaccoppiato.
- **Configurabile e Sicuro**: Tutta la configurazione avviene tramite variabili d'ambiente.

## 🚀 Come Iniziare

### Prerequisiti

- Docker e Docker Compose
- Un client HTTP come `curl` o Postman

### 1. Configurazione

Copia il file `.env.example` in un nuovo file chiamato `.env` e personalizza i valori per il tuo Identity Provider (IdP).

```bash
cp .env.example .env
```

**Configura almeno queste variabili nel file `.env`:**
- `JWKS_URL`: L'URL del tuo provider che espone le chiavi JWKS.
- `ALLOWED_ISSUERS`: L'issuer che ti aspetti di trovare nel token.
- `ALLOWED_AUDIENCES`: L'audience che ti aspetti di trovare nel token.

### 2. Esecuzione

Avvia l'intero stack (Traefik, servizio di auth e app di esempio) con un singolo comando:

```bash
docker-compose up --build
```

### 3. Test

Per testare, hai bisogno di un token JWT valido generato dal tuo Identity Provider.

Esporta il token generato in una variabile di shell:

```bash
export TOKEN="il_tuo_token_jwt_qui"
```

#### Testare l'accesso al servizio protetto

```bash
# 1. Richiesta senza token (verrà bloccata da Traefik) -> Forbidden
curl -H "Host: whoami.localhost" http://localhost

# 2. Richiesta con token valido -> OK
# La risposta includerà l'header "X-User-Id: <subject_dal_tuo_token>"
curl -H "Host: whoami.localhost" -H "Authorization: Bearer $TOKEN" http://localhost

# 3. Richiesta con un token non valido (es. scaduto o con issuer sbagliato) -> Forbidden
curl -H "Host: whoami.localhost" -H "Authorization: Bearer $INVALID_TOKEN" http://localhost
```