package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/franko/crypto-oracle/pkg/utils/config"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// Redis encapsule le client Redis et fournit des méthodes pour interagir avec le cache
type Redis struct {
	client *redis.Client
	ctx    context.Context
	logger *logrus.Logger
}

// NewRedisConnection crée une nouvelle connexion Redis
func NewRedisConnection(cfg *config.RedisConfig, logger *logrus.Logger) (*Redis, error) {
	ctx := context.Background()

	// Créer le client Redis
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	// Vérifier la connexion
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("échec de la connexion à Redis: %w", err)
	}

	logger.Info("Connected to Redis", logrus.Fields{
		"host": cfg.Host,
		"port": cfg.Port,
	})

	return &Redis{
		client: client,
		ctx:    ctx,
		logger: logger,
	}, nil
}

// Close ferme la connexion à Redis
func (r *Redis) Close() error {
	return r.client.Close()
}

// Set stocke une valeur dans le cache
func (r *Redis) Set(key string, value string, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// Get récupère une valeur du cache
func (r *Redis) Get(key string) (string, error) {
	return r.client.Get(r.ctx, key).Result()
}

// Delete supprime une clé du cache
func (r *Redis) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// SetStruct stocke une structure sérialisée en JSON dans le cache
func (r *Redis) SetStruct(key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(r.ctx, key, value, expiration).Err()
}

// GetStruct récupère et désérialise une structure en JSON du cache
func (r *Redis) GetStruct(key string, value interface{}) error {
	return r.client.Get(r.ctx, key).Scan(value)
}

// Keys récupère les clés correspondant à un pattern
func (r *Redis) Keys(pattern string) ([]string, error) {
	return r.client.Keys(r.ctx, pattern).Result()
}

// Exists vérifie si une clé existe
func (r *Redis) Exists(key string) (bool, error) {
	val, err := r.client.Exists(r.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// TTL récupère le temps de vie restant d'une clé
func (r *Redis) TTL(key string) (time.Duration, error) {
	return r.client.TTL(r.ctx, key).Result()
}

// PurgePattern supprime toutes les clés correspondant à un pattern
func (r *Redis) PurgePattern(pattern string) error {
	keys, err := r.client.Keys(r.ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(r.ctx, keys...).Err()
}

// XAdd ajoute un message à un stream
func (r *Redis) XAdd(stream string, values map[string]interface{}) error {
	return r.client.XAdd(r.ctx, &redis.XAddArgs{
		Stream: stream,
		ID:     "*", // Auto-generate ID
		Values: values,
	}).Err()
}

// XGroupCreate crée un groupe de consommateurs pour un stream
func (r *Redis) XGroupCreate(stream, group string) error {
	// Vérifier si le stream existe, sinon le créer avec un message vide
	exists, err := r.Exists(stream)
	if err != nil {
		return err
	}

	if !exists {
		// Créer le stream avec un message vide
		err = r.XAdd(stream, map[string]interface{}{"init": "true"})
		if err != nil {
			return err
		}
	}

	err = r.client.XGroupCreate(r.ctx, stream, group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}

	return nil
}

// XAck acquitte un message dans un groupe de consommateurs
func (r *Redis) XAck(stream, group, messageID string) error {
	return r.client.XAck(r.ctx, stream, group, messageID).Err()
}

// XReadGroup lit des messages d'un stream
func (r *Redis) XReadGroup(stream, group, consumer string, count int, timeout time.Duration) ([]XMessage, error) {
	streams := []string{stream, ">"}
	result, err := r.client.XReadGroup(r.ctx, &redis.XReadGroupArgs{
		Group:    group,
		Consumer: consumer,
		Streams:  streams,
		Count:    int64(count),
		Block:    timeout,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []XMessage{}, nil
		}
		return nil, err
	}

	var messages []XMessage
	for _, s := range result {
		for _, m := range s.Messages {
			messages = append(messages, XMessage{
				ID:     m.ID,
				Values: m.Values,
			})
		}
	}

	return messages, nil
} 