package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/franko/crypto-oracle/internal/storage/cache"
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

	// Convertir en format Redis (map[string]interface{})
	values := map[string]interface{}{
		"id":        message.ID,
		"type":      message.Type,
		"timestamp": message.Timestamp.Format(time.RFC3339),
	}

	// Ajouter tous les champs du payload
	for k, v := range message.Payload {
		values[k] = v
	}

	// Publier dans Redis Stream
	err := p.cache.XAdd(streamName, values)
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
						ts, err := time.Parse(time.RFC3339, v.(string))
						if err == nil {
							message.Timestamp = ts
						}
					} else {
						message.Payload[k] = v
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

// GetRedisClient retourne le client Redis du pipeline
func (p *Pipeline) GetRedisClient() *cache.Redis {
	return p.cache
} 