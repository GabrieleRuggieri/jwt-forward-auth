# --- Build Stage ---
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copia i file del modulo e scarica le dipendenze
COPY go.mod go.sum ./
RUN go mod download

# Copia il codice sorgente
COPY . .

# Compila l'applicazione
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/server ./cmd/server

# --- Final Stage ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copia il binario compilato dallo stage di build
COPY --from=builder /app/server .

# Espone la porta
EXPOSE 8080

# Comando di avvio
CMD ["./server"]