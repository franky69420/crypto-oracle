package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/franko/crypto-oracle/pkg/models"
	"github.com/franko/crypto-oracle/pkg/utils/config"
	"go.uber.org/zap"
)

// Client gère la connexion au cache Redis
type Client struct {
	client  *redis.Client
	ctx     context.Context
	config  *config.RedisConfig
	logger *zap.Logger
}

// NewClient crée un nouveau client Redis
func NewClient(ctx context.Context, config *config.RedisConfig) (*Client, error) {
	// Configurer le client Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
		PoolSize: config.PoolSize,
	})

	// Vérifier la connexion
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("échec de la connexion à Redis: %w", err)
	}

	logger, _ := zap.NewProduction()

	return &Client{
		client: redisClient,
		ctx:    ctx,
		config: config,
		logger: logger,
	}, nil
}

// Close ferme la connexion Redis
func (c *Client) Close() {
	if c.client != nil {
		c.client.Close()
	}
}

// Get récupère une valeur du cache
func (c *Client) Get(key string) (string, error) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return "", errors.New("clé non trouvée")
	} else if err != nil {
		return "", err
	}
	return val, nil
}

// Set stocke une valeur dans le cache
func (c *Client) Set(key string, value string, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, value, expiration).Err()
}

// SetStruct stocke une structure dans le cache
func (c *Client) SetStruct(key string, obj interface{}, expiration time.Duration) error {
	// Sérialiser l'objet en JSON
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("échec de la sérialisation en JSON: %w", err)
	}

	// Stocker le JSON dans Redis
	err = c.client.Set(c.ctx, key, jsonBytes, expiration).Err()
	if err != nil {
		return fmt.Errorf("échec du stockage dans Redis: %w", err)
	}
	return nil
}

// GetStruct récupère une structure du cache
func (c *Client) GetStruct(key string, obj interface{}) error {
	// Récupérer le JSON du cache
	jsonBytes, err := c.client.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		return errors.New("clé non trouvée")
	} else if err != nil {
		return err
	}

	// Désérialiser le JSON dans l'objet fourni
	if err := json.Unmarshal(jsonBytes, obj); err != nil {
		return fmt.Errorf("échec de la désérialisation du JSON: %w", err)
	}
	return nil
}

// Unmarshal désérialise une chaîne JSON en structure
func (c *Client) Unmarshal(data string, obj interface{}) error {
	if err := json.Unmarshal([]byte(data), obj); err != nil {
		return fmt.Errorf("échec de la désérialisation du JSON: %w", err)
	}
	return nil
}

// GetFloat64 récupère une valeur float64 du cache
func (c *Client) GetFloat64(key string) (float64, error) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return 0, errors.New("clé non trouvée")
	} else if err != nil {
		return 0, err
	}
	
	return strconv.ParseFloat(val, 64)
}

// SetFloat64 stocke une valeur float64 dans le cache
func (c *Client) SetFloat64(key string, value float64, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, strconv.FormatFloat(value, 'f', -1, 64), expiration).Err()
}

// Delete supprime une clé du cache
func (c *Client) Delete(key string) error {
	return c.client.Del(c.ctx, key).Err()
}

// Keys récupère les clés correspondant à un pattern
func (c *Client) Keys(pattern string) ([]string, error) {
	return c.client.Keys(c.ctx, pattern).Result()
}

// PurgePattern supprime toutes les clés correspondant à un pattern
func (c *Client) PurgePattern(pattern string) error {
	// Récupérer toutes les clés correspondant au pattern
	keys, err := c.client.Keys(c.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("échec de la récupération des clés: %w", err)
	}
	
	// Si aucune clé ne correspond, rien à faire
	if len(keys) == 0 {
		return nil
	}
	
	// Supprimer toutes les clés
	if err := c.client.Del(c.ctx, keys...).Err(); err != nil {
		return fmt.Errorf("échec de la suppression des clés: %w", err)
	}
	
	return nil
}

// TTL récupère le temps restant avant expiration d'une clé
func (c *Client) TTL(key string) (time.Duration, error) {
	return c.client.TTL(c.ctx, key).Result()
}

// Exists vérifie si une clé existe
func (c *Client) Exists(key string) (bool, error) {
	result, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Pipeline crée un pipeline Redis pour exécuter plusieurs commandes en une fois
func (c *Client) Pipeline() redis.Pipeliner {
	return c.client.Pipeline()
}

// LPush ajoute une valeur au début d'une liste
func (c *Client) LPush(key string, value interface{}) error {
	return c.client.LPush(c.ctx, key, value).Err()
}

// RPush ajoute une valeur à la fin d'une liste
func (c *Client) RPush(key string, value interface{}) error {
	return c.client.RPush(c.ctx, key, value).Err()
}

// LRange récupère une plage d'éléments d'une liste
func (c *Client) LRange(key string, start, stop int64) ([]string, error) {
	return c.client.LRange(c.ctx, key, start, stop).Result()
}

// HSet définit un champ dans un hash
func (c *Client) HSet(key, field string, value interface{}) error {
	return c.client.HSet(c.ctx, key, field, value).Err()
}

// HGet récupère un champ d'un hash
func (c *Client) HGet(key, field string) (string, error) {
	return c.client.HGet(c.ctx, key, field).Result()
}

// HGetAll récupère tous les champs d'un hash
func (c *Client) HGetAll(key string) (map[string]string, error) {
	return c.client.HGetAll(c.ctx, key).Result()
}

// Publish publie un message sur un canal
func (c *Client) Publish(channel, message string) error {
	return c.client.Publish(c.ctx, channel, message).Err()
}

// Subscribe s'abonne à un canal pour recevoir des messages
func (c *Client) Subscribe(channel string) *redis.PubSub {
	return c.client.Subscribe(c.ctx, channel)
}

// Expire définit un délai d'expiration pour une clé
func (c *Client) Expire(key string, expiration time.Duration) error {
	return c.client.Expire(c.ctx, key, expiration).Err()
}

// Incr incrémente la valeur d'une clé
func (c *Client) Incr(key string) (int64, error) {
	return c.client.Incr(c.ctx, key).Result()
}

// ZAdd ajoute un membre à un ensemble trié
func (c *Client) ZAdd(key string, score float64, member string) error {
	return c.client.ZAdd(c.ctx, key, &redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// ZRange récupère une plage de membres d'un ensemble trié
func (c *Client) ZRange(key string, start, stop int64) ([]string, error) {
	return c.client.ZRange(c.ctx, key, start, stop).Result()
}

// ZRangeByScore récupère des membres d'un ensemble trié par score
func (c *Client) ZRangeByScore(key string, min, max float64) ([]string, error) {
	return c.client.ZRangeByScore(c.ctx, key, &redis.ZRangeBy{
		Min: strconv.FormatFloat(min, 'f', -1, 64),
		Max: strconv.FormatFloat(max, 'f', -1, 64),
	}).Result()
}

// CacheActiveWallets met en cache une liste de wallets actifs
func (c *Client) CacheActiveWallets(tokenAddress string, wallets []models.ActiveWallet, expiration time.Duration) error {
	key := fmt.Sprintf("token:%s:active_wallets", tokenAddress)
	return c.SetStruct(key, wallets, expiration)
}

// GetCachedActiveWallets récupère les wallets actifs depuis le cache
func (c *Client) GetCachedActiveWallets(tokenAddress string) ([]models.ActiveWallet, error) {
	key := fmt.Sprintf("token:%s:active_wallets", tokenAddress)
	var wallets []models.ActiveWallet
	err := c.GetStruct(key, &wallets)
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

// CacheActiveWalletsByTrustScore met en cache une liste de wallets actifs triés par score de confiance
func (c *Client) CacheActiveWalletsByTrustScore(tokenAddress string, minScore float64, wallets []models.ActiveWallet, expiration time.Duration) error {
	key := fmt.Sprintf("token:%s:active_wallets:trust_score:%f", tokenAddress, minScore)
	return c.SetStruct(key, wallets, expiration)
}

// GetCachedActiveWalletsByTrustScore récupère les wallets actifs triés par score de confiance depuis le cache
func (c *Client) GetCachedActiveWalletsByTrustScore(tokenAddress string, minScore float64) ([]models.ActiveWallet, error) {
	key := fmt.Sprintf("token:%s:active_wallets:trust_score:%f", tokenAddress, minScore)
	var wallets []models.ActiveWallet
	err := c.GetStruct(key, &wallets)
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

// CacheActiveWalletsCount met en cache le nombre de wallets actifs sur un token
func (c *Client) CacheActiveWalletsCount(tokenAddress string, count int, expiration time.Duration) error {
	key := fmt.Sprintf("token:%s:active_wallets:count", tokenAddress)
	return c.SetInt(key, count, expiration)
}

// GetCachedActiveWalletsCount récupère le nombre de wallets actifs depuis le cache
func (c *Client) GetCachedActiveWalletsCount(tokenAddress string) (int, error) {
	key := fmt.Sprintf("token:%s:active_wallets:count", tokenAddress)
	return c.GetInt(key)
}

// GetInt récupère une valeur entière du cache
func (c *Client) GetInt(key string) (int, error) {
	val, err := c.client.Get(c.ctx, key).Result()
	if err == redis.Nil {
		return 0, errors.New("clé non trouvée")
	} else if err != nil {
		return 0, err
	}
	
	return strconv.Atoi(val)
}

// SetInt stocke une valeur entière dans le cache
func (c *Client) SetInt(key string, value int, expiration time.Duration) error {
	return c.client.Set(c.ctx, key, strconv.Itoa(value), expiration).Err()
}

// XAdd ajoute un message à un stream Redis
func (c *Client) XAdd(stream string, values map[string]interface{}) error {
	// Convertir les valeurs complexes en string JSON
	processedValues := make(map[string]interface{})
	for k, v := range values {
		switch val := v.(type) {
		case map[string]interface{}, []interface{}:
			// Sérialiser les structures complexes en JSON
			jsonBytes, err := json.Marshal(val)
			if err != nil {
				return fmt.Errorf("marshaling error: %w", err)
			}
			processedValues[k] = string(jsonBytes)
		default:
			// Garder les valeurs simples telles quelles
			processedValues[k] = val
		}
	}

	// Convertir les valeurs en Args compatibles avec go-redis
	args := make([]interface{}, 0, len(processedValues)*2+3)
	args = append(args, "XADD", stream, "*") // "*" génère un ID automatique

	for k, v := range processedValues {
		args = append(args, k, v)
	}

	// Exécuter la commande XADD
	cmd := redis.NewStringCmd(c.ctx, args...)
	c.client.Process(c.ctx, cmd)
	_, err := cmd.Result()
	return err
}

// XMessage représente un message dans un Redis Stream
type XMessage struct {
	ID     string
	Values map[string]interface{}
}

// XReadGroup lit des messages d'un stream avec un groupe de consommateurs
func (c *Client) XReadGroup(stream, group, consumer string, count int, timeout time.Duration) ([]XMessage, error) {
	// Préparer les arguments pour XREADGROUP
	args := []interface{}{
		"XREADGROUP", "GROUP", group, consumer,
		"COUNT", count,
		"BLOCK", int64(timeout / time.Millisecond),
		"STREAMS", stream, ">", // ">" signifie nouveaux messages uniquement
	}

	// Exécuter la commande XREADGROUP
	cmd := redis.NewCmd(c.ctx, args...)
	c.client.Process(c.ctx, cmd)
	result, err := cmd.Result()
	if err != nil {
		if err == redis.Nil {
			// Pas de nouveaux messages
			return []XMessage{}, nil
		}
		return nil, err
	}

	// Traiter le résultat
	streams, ok := result.([]interface{})
	if !ok || len(streams) == 0 {
		return []XMessage{}, nil
	}

	// Extraire le nom du stream et les messages
	streamData, ok := streams[0].([]interface{})
	if !ok || len(streamData) < 2 {
		return []XMessage{}, nil
	}

	// Récupérer les messages
	messages, ok := streamData[1].([]interface{})
	if !ok {
		return []XMessage{}, nil
	}

	// Convertir les messages en XMessage
	var xMessages []XMessage
	for _, msgData := range messages {
		msgArray, ok := msgData.([]interface{})
		if !ok || len(msgArray) < 2 {
			continue
		}

		// ID du message
		id, ok := msgArray[0].(string)
		if !ok {
			continue
		}

		// Valeurs du message
		fieldsArray, ok := msgArray[1].([]interface{})
		if !ok || len(fieldsArray)%2 != 0 {
			continue
		}

		// Construire la map de valeurs
		values := make(map[string]interface{})
		for i := 0; i < len(fieldsArray); i += 2 {
			fieldName, ok := fieldsArray[i].(string)
			if !ok {
				continue
			}
			values[fieldName] = fieldsArray[i+1]
		}

		xMessages = append(xMessages, XMessage{
			ID:     id,
			Values: values,
		})
	}

	return xMessages, nil
}

// XGroupCreate crée un groupe de consommateurs pour un stream
func (c *Client) XGroupCreate(stream, group string) error {
	// Vérifier si le stream existe, sinon le créer avec un message vide
	exists, err := c.Exists(stream)
	if err != nil {
		return err
	}

	if !exists {
		// Créer le stream avec un message vide
		err = c.XAdd(stream, map[string]interface{}{"init": "true"})
		if err != nil {
			return err
		}
	}

	// Créer le groupe de consommateurs
	cmd := redis.NewStatusCmd(c.ctx, "XGROUP", "CREATE", stream, group, "$", "MKSTREAM")
	c.client.Process(c.ctx, cmd)
	_, err = cmd.Result()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	// Si l'erreur est BUSYGROUP, le groupe existe déjà
	if err != nil && strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf("BUSYGROUP Consumer Group name already exists")
	}

	return nil
}

// XAck acquitte un message dans un groupe de consommateurs
func (c *Client) XAck(stream, group, messageID string) error {
	cmd := redis.NewIntCmd(c.ctx, "XACK", stream, group, messageID)
	c.client.Process(c.ctx, cmd)
	_, err := cmd.Result()
	return err
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
} 