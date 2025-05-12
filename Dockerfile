FROM golang:1.21-alpine AS builder

# Installer les dépendances de compilation
RUN apk add --no-cache git build-base tzdata ca-certificates

# Définir le répertoire de travail
WORKDIR /src

# Copier les fichiers de dépendances
COPY go.mod go.sum ./
RUN go mod download

# Copier le reste du code source
COPY . .

# Configurer la compilation pour une exécution en production sans debug
RUN go build -ldflags="-s -w" -o /app/crypto-oracle ./cmd/main.go

# Étape 2: Image finale légère
FROM alpine:3.18

# Installer les dépendances nécessaires
RUN apk add --no-cache ca-certificates tzdata curl

# Créer un utilisateur non-root
RUN addgroup -S app && adduser -S -G app app

# Créer les répertoires nécessaires
RUN mkdir -p /app/config /app/logs && \
    chown -R app:app /app

# Définir le répertoire de travail
WORKDIR /app

# Copier l'exécutable depuis l'étape précédente
COPY --from=builder --chown=app:app /app/crypto-oracle .

# Copier les fichiers de configuration
COPY --chown=app:app config/prod.yaml /app/config/prod.yaml

# Exposition des ports
EXPOSE 3000

# Changer l'utilisateur
USER app

# Définir les variables d'environnement
ENV CONFIG_FILE=/app/config/prod.yaml

# Commande d'entrée
CMD ["./crypto-oracle", "-config", "/app/config/prod.yaml"] 