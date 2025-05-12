# Crypto Oracle - Memecoin Detector v4.2

Crypto Oracle est un système d'analyse et de détection de memecoins basé sur des métriques avancées et un réseau de confiance de wallets.

## Fonctionnalités

- **Alert Manager**: Système d'alertes pour détecter des événements importants sur les tokens
- **Pipeline Service**: Infrastructure de traitement des événements en temps réel
- **Reactivation System**: Détection de la réactivation de tokens dormants
- **Token Engine**: Moteur d'analyse des tokens et de leurs métriques
- **Memory of Trust**: Réseau de confiance pour évaluer la fiabilité des wallets
- **Wallet Intelligence**: Analyse avancée des comportements des wallets

## Architecture

Le système est composé de plusieurs modules:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Alert Manager │    │ Pipeline Service │    │  API Service    │
└────────┬────────┘    └────────┬────────┘    └────────┬────────┘
         │                      │                      │
         └──────────────┬───────┴──────────────┬──────┘
                        │                      │
               ┌────────┴────────┐    ┌────────┴────────┐
               │   Token Engine  │    │Wallet Intelligence
               └────────┬────────┘    └────────┬────────┘
                        │                      │
                        └──────────┬───────────┘
                                   │
                        ┌──────────┴───────────┐
                        │   Memory of Trust    │
                        └──────────┬───────────┘
                                   │
                        ┌──────────┴───────────┐
                        │   Data Storage       │
                        │  (PostgreSQL/Redis)  │
                        └──────────────────────┘
```

## Technologies utilisées

- **Go**: Langage de programmation principal
- **PostgreSQL**: Base de données principale
- **Redis**: Cache et pipeline de messages
- **Docker**: Conteneurisation pour le déploiement
- **Prometheus/Grafana**: Surveillance et métriques

## Installation

### Prérequis

- Go 1.20 ou supérieur
- Docker et Docker Compose
- PostgreSQL 13+ et Redis 6+

### Avec Docker (recommandé)

1. Cloner le dépôt:
   ```bash
   git clone https://github.com/franky69420/crypto-oracle.git
   cd crypto-oracle
   ```

2. Configurer les variables d'environnement:
   ```bash
   cp .env.example .env
   # Éditer le fichier .env avec vos paramètres
   ```

3. Démarrer les services:
   ```bash
   # Environnement de développement
   docker-compose up -d
   
   # Environnement de production
   docker-compose -f docker-compose.prod.yml up -d
   ```

### Développement local

1. Démarrer les services de base de données:
   ```bash
   docker-compose up -d postgres redis
   ```

2. Compiler et exécuter l'application:
   ```bash
   go build -o crypto-oracle ./cmd/main.go
   ./crypto-oracle -config config/dev.yaml
   ```

## Tests

### Exécuter les tests unitaires

```bash
go test ./...
```

### Exécuter les tests d'intégration

```bash
go test -tags=integration ./...
```

### Tester des modules spécifiques

```bash
# Tester le Pipeline et le Token Engine
go run test-pipeline-token.go
```

## Déploiement en production

Le projet inclut une configuration Docker Compose complète pour un déploiement en production:

```bash
# Préparation
mkdir -p ./monitoring/prometheus ./monitoring/grafana/provisioning ./traefik

# Configuration
cp config/prod.example.yaml config/prod.yaml
# Éditer config/prod.yaml selon vos besoins

# Démarrage
docker-compose -f docker-compose.prod.yml up -d
```

La configuration de production inclut:
- Traefik pour le routage HTTPS avec Let's Encrypt
- Prometheus et Grafana pour la surveillance
- Répliques et volumes persistants pour les données

## API Endpoints

L'API REST est disponible sur le port 3000 par défaut:

- `GET /api/health`: Vérification de l'état du service
- `GET /api/tokens/{token_id}`: Informations sur un token
- `GET /api/tokens/{token_id}/active-wallets`: Liste des wallets actifs sur un token

## Contribuer

Les contributions sont les bienvenues! Veuillez suivre ces étapes:

1. Forker le dépôt
2. Créer une branche (`git checkout -b feature/amazing-feature`)
3. Committer vos changements (`git commit -m 'Add some amazing feature'`)
4. Pousser vers la branche (`git push origin feature/amazing-feature`)
5. Ouvrir une Pull Request

## Licence

Distribué sous la licence MIT. Voir `LICENSE` pour plus d'informations.

## Contact

Franky - [@franky69420](https://github.com/franky69420) 