package api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// RedisHandler handles Redis-related API endpoints
type RedisHandler struct {
	redisClient *redis.Client
}

// NewRedisHandler creates a new Redis handler
func NewRedisHandler(redisAddr string, isUnixSocket bool) *RedisHandler {
	// Determine network type
	networkType := "tcp"
	if isUnixSocket {
		networkType = "unix"
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Network:      networkType,
		Addr:         redisAddr,
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	return &RedisHandler{
		redisClient: client,
	}
}

// RegisterRoutes registers Redis routes to the fiber app
func (h *RedisHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/api/redis")

	// @Summary Set a Redis key
	// @Description Set a key-value pair in Redis with optional expiration
	// @Tags redis
	// @Accept json
	// @Produce json
	// @Param request body SetKeyRequest true "Key-value data"
	// @Success 200 {object} SetKeyResponse
	// @Failure 400 {object} ErrorResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/set [post]
	group.Post("/set", h.setKey)

	// @Summary Get a Redis key
	// @Description Get a value by key from Redis
	// @Tags redis
	// @Produce json
	// @Param key path string true "Key to retrieve"
	// @Success 200 {object} GetKeyResponse
	// @Failure 400 {object} ErrorResponse
	// @Failure 404 {object} ErrorResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/get/{key} [get]
	group.Get("/get/:key", h.getKey)

	// @Summary Delete a Redis key
	// @Description Delete a key from Redis
	// @Tags redis
	// @Produce json
	// @Param key path string true "Key to delete"
	// @Success 200 {object} DeleteKeyResponse
	// @Failure 400 {object} ErrorResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/del/{key} [delete]
	group.Delete("/del/:key", h.deleteKey)

	// @Summary Get Redis keys by pattern
	// @Description Get keys matching a pattern from Redis
	// @Tags redis
	// @Produce json
	// @Param pattern path string true "Pattern to match keys"
	// @Success 200 {object} GetKeysResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/keys/{pattern} [get]
	group.Get("/keys/:pattern", h.getKeys)

	// @Summary Set hash fields
	// @Description Set one or more fields in a Redis hash
	// @Tags redis
	// @Accept json
	// @Produce json
	// @Param request body HSetKeyRequest true "Hash field data"
	// @Success 200 {object} HSetKeyResponse
	// @Failure 400 {object} ErrorResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/hset [post]
	group.Post("/hset", h.hsetKey)

	// @Summary Get hash field
	// @Description Get a field from a Redis hash
	// @Tags redis
	// @Produce json
	// @Param key path string true "Hash key"
	// @Param field path string true "Field to retrieve"
	// @Success 200 {object} HGetKeyResponse
	// @Failure 400 {object} ErrorResponse
	// @Failure 404 {object} ErrorResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/hget/{key}/{field} [get]
	group.Get("/hget/:key/:field", h.hgetKey)

	// @Summary Delete hash fields
	// @Description Delete one or more fields from a Redis hash
	// @Tags redis
	// @Accept json
	// @Produce json
	// @Param request body HDelKeyRequest true "Fields to delete"
	// @Success 200 {object} HDelKeyResponse
	// @Failure 400 {object} ErrorResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/hdel [post]
	group.Post("/hdel", h.hdelKey)

	// @Summary Get hash fields
	// @Description Get all field names in a Redis hash
	// @Tags redis
	// @Produce json
	// @Param key path string true "Hash key"
	// @Success 200 {object} HKeysResponse
	// @Failure 400 {object} ErrorResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/hkeys/{key} [get]
	group.Get("/hkeys/:key", h.hkeysKey)

	// @Summary Get all hash fields and values
	// @Description Get all fields and values in a Redis hash
	// @Tags redis
	// @Produce json
	// @Param key path string true "Hash key"
	// @Success 200 {object} map[string]string
	// @Failure 400 {object} ErrorResponse
	// @Failure 500 {object} ErrorResponse
	// @Router /api/redis/hgetall/{key} [get]
	group.Get("/hgetall/:key", h.hgetallKey)
}

// setKey sets a key-value pair in Redis
func (h *RedisHandler) setKey(c *fiber.Ctx) error {
	// Parse request
	var req struct {
		Key     string `json:"key"`
		Value   string `json:"value"`
		Expires int    `json:"expires,omitempty"` // Expiration in seconds, optional
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
	}

	// Validate required fields
	if req.Key == "" || req.Value == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Key and value are required",
		})
	}

	ctx := context.Background()
	var err error

	// Set with or without expiration
	if req.Expires > 0 {
		err = h.redisClient.Set(ctx, req.Key, req.Value, time.Duration(req.Expires)*time.Second).Err()
	} else {
		err = h.redisClient.Set(ctx, req.Key, req.Value, 0).Err()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to set key: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Key set successfully",
	})
}

// getKey retrieves a value by key from Redis
func (h *RedisHandler) getKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Key is required",
		})
	}

	ctx := context.Background()
	val, err := h.redisClient.Get(ctx, key).Result()

	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Key not found",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get key: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"key":     key,
		"value":   val,
	})
}

// deleteKey deletes a key from Redis
func (h *RedisHandler) deleteKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Key is required",
		})
	}

	ctx := context.Background()
	result, err := h.redisClient.Del(ctx, key).Result()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete key: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"deleted": result > 0,
		"count":   result,
	})
}

// getKeys retrieves keys matching a pattern from Redis
func (h *RedisHandler) getKeys(c *fiber.Ctx) error {
	pattern := c.Params("pattern", "*")

	ctx := context.Background()
	keys, err := h.redisClient.Keys(ctx, pattern).Result()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get keys: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"keys":    keys,
		"count":   len(keys),
	})
}

// hsetKey sets a field in a hash stored at key
func (h *RedisHandler) hsetKey(c *fiber.Ctx) error {
	// Parse request
	var req struct {
		Key    string            `json:"key"`
		Fields map[string]string `json:"fields"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
	}

	// Validate required fields
	if req.Key == "" || len(req.Fields) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Key and at least one field are required",
		})
	}

	ctx := context.Background()
	totalAdded := 0

	// Use HSet to set multiple fields at once
	for field, value := range req.Fields {
		added, err := h.redisClient.HSet(ctx, req.Key, field, value).Result()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to set hash field: " + err.Error(),
			})
		}
		totalAdded += int(added)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"added":   totalAdded,
	})
}

// hgetKey retrieves a field from a hash stored at key
func (h *RedisHandler) hgetKey(c *fiber.Ctx) error {
	key := c.Params("key")
	field := c.Params("field")

	if key == "" || field == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Key and field are required",
		})
	}

	ctx := context.Background()
	val, err := h.redisClient.HGet(ctx, key, field).Result()

	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "Field not found in hash",
		})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get hash field: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"key":     key,
		"field":   field,
		"value":   val,
	})
}

// hdelKey deletes fields from a hash stored at key
func (h *RedisHandler) hdelKey(c *fiber.Ctx) error {
	// Parse request
	var req struct {
		Key    string   `json:"key"`
		Fields []string `json:"fields"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
	}

	// Validate required fields
	if req.Key == "" || len(req.Fields) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Key and at least one field are required",
		})
	}

	ctx := context.Background()
	fields := make([]string, len(req.Fields))
	copy(fields, req.Fields)

	removed, err := h.redisClient.HDel(ctx, req.Key, fields...).Result()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to delete hash fields: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"removed": removed,
	})
}

// hkeysKey retrieves all field names in a hash stored at key
func (h *RedisHandler) hkeysKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Key is required",
		})
	}

	ctx := context.Background()
	fields, err := h.redisClient.HKeys(ctx, key).Result()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get hash keys: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"key":     key,
		"fields":  fields,
		"count":   len(fields),
	})
}

// hgetallKey retrieves all fields and values in a hash stored at key
func (h *RedisHandler) hgetallKey(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Key is required",
		})
	}

	ctx := context.Background()
	values, err := h.redisClient.HGetAll(ctx, key).Result()

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to get hash: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"key":     key,
		"hash":    values,
		"count":   len(values),
	})
}
