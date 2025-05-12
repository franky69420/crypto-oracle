FROM golang:1.21-alpine AS builder

# Installer les dépendances du système
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Définir le répertoire de travail
WORKDIR /app

# Copier les fichiers go.mod et go.sum
COPY go.mod go.sum ./

# Télécharger les dépendances
RUN go mod download

# Copier le code source
COPY . .

# Construire l'application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crypto-oracle ./cmd/main.go

# Image finale
FROM alpine:3.18

# Installer les dépendances du système
RUN apk add --no-cache ca-certificates tzdata

# Copier l'application depuis l'étape de build
COPY --from=builder /app/crypto-oracle /app/crypto-oracle

# Copier les configurations
COPY config.yaml /app/config.yaml

# Définir les variables d'environnement
ENV GIN_MODE=release

# Exposer le port
EXPOSE 8080

# Définir le répertoire de travail
WORKDIR /app

# Commande par défaut
CMD ["/app/crypto-oracle"] 