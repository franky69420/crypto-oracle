package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/franky69420/crypto-oracle/internal/storage/cache"
	"github.com/franky69420/crypto-oracle/pkg/models"
	"github.com/sirupsen/logrus"
)

// Pipeline gère les flux de traitement des données
type Pipeline struct {
	cache      *cache.Redis
	logger     *logrus.Logger
	processors map[string]Processor
	stopped    bool
}

// Processor est une interface pour les processeurs de messages
type Processor interface {
	Process(message Message) error
	GetName() string
}

// Message représente un message à traiter dans le pipeline
type Message struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Payload   map[string]interface{} `json:"payload"`
}

// NewPipeline crée un nouveau pipeline
func NewPipeline(cache *cache.Redis, logger *logrus.Logger) *Pipeline {
	return &Pipeline{
		cache:      cache,
		logger:     logger,
		processors: make(map[string]Processor),
		stopped:    true,
	}
}

// Start démarre le pipeline
func (p *Pipeline) Start(ctx context.Context) error {
	p.logger.Info("Starting Pipeline")
	p.stopped = false

	// Démarrer les goroutines de consommation pour chaque processeur
	for name, processor := range p.processors {
		go p.startConsumer(ctx, name, processor)
	}

	return nil
}

// Shutdown arrête le pipeline
func (p *Pipeline) Shutdown(ctx context.Context) error {
	p.logger.Info("Shutting down Pipeline")
	p.stopped = true
	// Attendre que les goroutines se terminent
	time.Sleep(500 * time.Millisecond)
	return nil
}

// RegisterProcessor enregistre un processeur de messages
func (p *Pipeline) RegisterProcessor(processor Processor) {
	p.processors[processor.GetName()] = processor
	p.logger.WithFields(logrus.Fields{
		"processor": processor.GetName(),
	}).Info("Processor registered")
}

// PublishMessage publie un message dans un stream
func (p *Pipeline) PublishMessage(streamName string, message Message) error {
	// Ajouter un ID si non fourni
	if message.ID == "" {
		message.ID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}

	// Ajouter timestamp si non fourni
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	// Sérialiser tout le message en JSON d'abord
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Désérialiser en map pour Redis
	var messageMap map[string]interface{}
	if err := json.Unmarshal(jsonData, &messageMap); err != nil {
		return fmt.Errorf("failed to unmarshal message to map: %w", err)
	}

	// Traiter spécifiquement les structures complexes
	for k, v := range messageMap {
		// Si c'est une structure complexe (map ou slice), la resérialiser en JSON
		switch val := v.(type) {
		case map[string]interface{}, []interface{}:
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("failed to marshal nested data: %w", err)
			}
			messageMap[k] = string(jsonBytes)
		}
	}

	// Publier dans Redis Stream
	err = p.cache.XAdd(streamName, messageMap)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"stream": streamName,
		"msg_id": message.ID,
		"type":   message.Type,
	}).Debug("Message published")

	return nil
}

// startConsumer démarre un consumer pour un processeur spécifique
func (p *Pipeline) startConsumer(ctx context.Context, streamName string, processor Processor) {
	p.logger.WithFields(logrus.Fields{
		"stream":    streamName,
		"processor": processor.GetName(),
	}).Info("Starting consumer")

	// Créer un consumer group si n'existe pas
	err := p.cache.XGroupCreate(streamName, processor.GetName())
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		p.logger.WithFields(logrus.Fields{
			"stream":    streamName,
			"processor": processor.GetName(),
			"error":     err.Error(),
		}).Error("Failed to create consumer group")
		return
	}

	// Boucle de consommation
	for !p.stopped {
		select {
		case <-ctx.Done():
			return
		default:
			// Lire les messages
			messages, err := p.cache.XReadGroup(streamName, processor.GetName(), "consumer1", 10, 1*time.Second)
			if err != nil {
				// Ignorer les timeouts (cas normal quand pas de messages)
				if err.Error() != "redis: nil" {
					p.logger.WithFields(logrus.Fields{
						"stream":    streamName,
						"processor": processor.GetName(),
						"error":     err.Error(),
					}).Error("Error reading from stream")
				}
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Traiter chaque message
			for _, msg := range messages {
				message := Message{
					ID:        msg.ID,
					Timestamp: time.Now(),
					Payload:   make(map[string]interface{}),
				}

				// Extraire les champs du message
				for k, v := range msg.Values {
					if k == "type" {
						message.Type = v.(string)
					} else if k == "timestamp" {
						// Détecter si c'est un timestamp sous forme de string ou un unix timestamp
						switch tv := v.(type) {
						case string:
							ts, err := time.Parse(time.RFC3339, tv)
							if err == nil {
								message.Timestamp = ts
							}
						case float64:
							message.Timestamp = time.Unix(int64(tv), 0)
						}
					} else {
						// Pour les autres champs, vérifier si c'est du JSON sérialisé
						if strVal, ok := v.(string); ok && (strings.HasPrefix(strVal, "{") || strings.HasPrefix(strVal, "[")) {
							// Tenter de désérialiser le JSON
							var obj interface{}
							if err := json.Unmarshal([]byte(strVal), &obj); err == nil {
								// Si c'est bien du JSON, l'ajouter tel quel
								message.Payload[k] = obj
							} else {
								// Sinon ajouter comme string
								message.Payload[k] = strVal
							}
						} else {
							// Ajouter directement si ce n'est pas un JSON sérialisé
							message.Payload[k] = v
						}
					}
				}

				// Traiter le message
				err := processor.Process(message)
				if err != nil {
					p.logger.WithFields(logrus.Fields{
						"stream":    streamName,
						"processor": processor.GetName(),
						"msg_id":    msg.ID,
						"error":     err.Error(),
					}).Error("Error processing message")
					// Ne pas ACK, sera retraité
					continue
				}

				// ACK si traité avec succès
				err = p.cache.XAck(streamName, processor.GetName(), msg.ID)
				if err != nil {
					p.logger.WithFields(logrus.Fields{
						"stream":    streamName,
						"processor": processor.GetName(),
						"msg_id":    msg.ID,
						"error":     err.Error(),
					}).Error("Error acknowledging message")
				}
			}
		}
	}
}

// TokenDetectionProcessor est un processeur pour les détections de tokens
type TokenDetectionProcessor struct {
	name string
}

// NewTokenDetectionProcessor crée un nouveau processeur de détection de tokens
func NewTokenDetectionProcessor() *TokenDetectionProcessor {
	return &TokenDetectionProcessor{
		name: "token_detection",
	}
}

// Process traite un message de détection de token
func (p *TokenDetectionProcessor) Process(message Message) error {
	// Logique de traitement à implémenter
	return nil
}

// GetName retourne le nom du processeur
func (p *TokenDetectionProcessor) GetName() string {
	return p.name
}

// TokenProcessor est un processeur pour les événements de tokens
type TokenProcessor struct {
	name        string
	tokenEngine interface {
		GetToken(tokenAddress string) (*models.Token, error)
		UpdateTokenState(tokenAddress, newState string) error
		SaveReactivationMetrics(candidate models.ReactivationCandidate) error
	}
	logger      *logrus.Logger
}

// NewTokenProcessor crée un nouveau processeur d'événements de tokens
func NewTokenProcessor(tokenEngine interface {
	GetToken(tokenAddress string) (*models.Token, error)
	UpdateTokenState(tokenAddress, newState string) error
	SaveReactivationMetrics(candidate models.ReactivationCandidate) error
}, logger *logrus.Logger) *TokenProcessor {
	return &TokenProcessor{
		name:        "token_processor",
		tokenEngine: tokenEngine,
		logger:      logger,
	}
}

// Process traite un événement de token
func (p *TokenProcessor) Process(message Message) error {
	p.logger.WithFields(logrus.Fields{
		"msg_id":   message.ID,
		"msg_type": message.Type,
	}).Debug("Processing token event")

	// Valider le message
	tokenAddress, ok := message.Payload["token_address"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid token_address in message payload")
	}

	// Traiter différents types d'événements
	switch message.Type {
	case "price_change":
		return p.processPriceChange(tokenAddress, message.Payload)
	case "volume_spike":
		return p.processVolumeSpike(tokenAddress, message.Payload)
	case "reactivation":
		return p.processReactivation(tokenAddress, message.Payload)
	case "state_change":
		return p.processStateChange(tokenAddress, message.Payload)
	default:
		p.logger.WithFields(logrus.Fields{
			"event_type": message.Type,
		}).Info("Unknown event type, ignoring")
		return nil
	}
}

// processPriceChange traite un changement de prix
func (p *TokenProcessor) processPriceChange(tokenAddress string, payload map[string]interface{}) error {
	// Extraire les données
	priceChange, ok := payload["price_change"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid price_change in payload")
	}

	p.logger.WithFields(logrus.Fields{
		"token_address": tokenAddress,
		"price_change":  priceChange,
	}).Info("Processing price change event")

	// Logique à implémenter selon les besoins
	return nil
}

// processVolumeSpike traite un pic de volume
func (p *TokenProcessor) processVolumeSpike(tokenAddress string, payload map[string]interface{}) error {
	// Extraire les données
	volume, ok := payload["volume"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid volume in payload")
	}

	p.logger.WithFields(logrus.Fields{
		"token_address": tokenAddress,
		"volume":        volume,
	}).Info("Processing volume spike event")

	// Logique à implémenter selon les besoins
	return nil
}

// processReactivation traite une réactivation de token
func (p *TokenProcessor) processReactivation(tokenAddress string, payload map[string]interface{}) error {
	// Extraire les données
	reactivationScore, ok := payload["reactivation_score"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid reactivation_score in payload")
	}

	p.logger.WithFields(logrus.Fields{
		"token_address":      tokenAddress,
		"reactivation_score": reactivationScore,
	}).Info("Processing token reactivation event")

	// Changer l'état du token
	err := p.tokenEngine.UpdateTokenState(tokenAddress, models.LifecycleStateReactivated)
	if err != nil {
		return fmt.Errorf("failed to update token state: %w", err)
	}

	return nil
}

// processStateChange traite un changement d'état de token
func (p *TokenProcessor) processStateChange(tokenAddress string, payload map[string]interface{}) error {
	// Extraire les données
	newState, ok := payload["new_state"].(string)
	if !ok {
		return fmt.Errorf("missing or invalid new_state in payload")
	}

	p.logger.WithFields(logrus.Fields{
		"token_address": tokenAddress,
		"new_state":     newState,
	}).Info("Processing token state change event")

	// Mettre à jour l'état du token
	err := p.tokenEngine.UpdateTokenState(tokenAddress, newState)
	if err != nil {
		return fmt.Errorf("failed to update token state: %w", err)
	}

	return nil
}

// GetName retourne le nom du processeur
func (p *TokenProcessor) GetName() string {
	return p.name
}

// GetRedisClient retourne le client Redis du pipeline
func (p *Pipeline) GetRedisClient() *cache.Redis {
	return p.cache
} 