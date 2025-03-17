package routes

import (
	"time"

	"github.com/freeflowuniverse/herolauncher/pkg/herolauncher/api"
	"github.com/freeflowuniverse/herolauncher/pkg/redisserver"
	"github.com/gofiber/fiber/v2"
)

// RedisHandler handles Redis-related API endpoints
type RedisHandler struct {
	redisServer *redisserver.Server
}

// NewRedisHandler creates a new Redis handler
func NewRedisHandler(server *redisserver.Server) *RedisHandler {
	return &RedisHandler{
		redisServer: server,
	}
}

// RegisterRoutes registers Redis routes to the fiber app
func (h *RedisHandler) RegisterRoutes(app *fiber.App) {
	group := app.Group("/api/redis")

	group.Post("/set", h.setKey)
	group.Get("/get/:key", h.getKey)
	group.Delete("/del/:key", h.deleteKey)
	group.Get("/keys/:pattern", h.getKeys)
	group.Post("/hset", h.hsetKey)
	group.Get("/hget/:key/:field", h.hgetKey)
	group.Post("/hdel", h.hdelKey)
	group.Get("/hkeys/:key", h.hkeysKey)
	group.Get("/hlen/:key", h.hlenKey)
	group.Post("/incr/:key", h.incrKey)
}

// @Summary Set a key
// @Description Set a key with a value and optional expiration
// @Tags redis
// @Accept json
// @Produce json
// @Param data body api.SetKeyRequest true "Key data"
// @Success 200 {object} api.SetKeyResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /api/redis/set [post]
func (h *RedisHandler) setKey(c *fiber.Ctx) error {
	var req api.SetKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	var duration time.Duration
	if req.ExpirationSeconds > 0 {
		duration = time.Duration(req.ExpirationSeconds) * time.Second
	}

	h.redisServer.Set(req.Key, req.Value, duration)

	return c.JSON(api.SetKeyResponse{
		Success: true,
	})
}

// @Summary Get a key
// @Description Get the value of a key
// @Tags redis
// @Produce json
// @Param key path string true "Key to get"
// @Success 200 {object} api.GetKeyResponse
// @Failure 404 {object} api.ErrorResponse
// @Router /api/redis/get/{key} [get]
func (h *RedisHandler) getKey(c *fiber.Ctx) error {
	key := c.Params("key")

	value, ok := h.redisServer.Get(key)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(api.ErrorResponse{
			Error: "Key not found",
		})
	}

	strValue, ok := value.(string)
	if !ok {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: "Value is not a string",
		})
	}

	return c.JSON(api.GetKeyResponse{
		Value: strValue,
	})
}

// @Summary Delete a key
// @Description Delete a key and return the number of keys removed
// @Tags redis
// @Produce json
// @Param key path string true "Key to delete"
// @Success 200 {object} api.DeleteKeyResponse
// @Router /api/redis/del/{key} [delete]
func (h *RedisHandler) deleteKey(c *fiber.Ctx) error {
	key := c.Params("key")

	count := h.redisServer.Del(key)

	return c.JSON(api.DeleteKeyResponse{
		Count: count,
	})
}

// @Summary Get keys matching a pattern
// @Description Get all keys matching the given pattern
// @Tags redis
// @Produce json
// @Param pattern path string true "Pattern to match"
// @Success 200 {object} api.GetKeysResponse
// @Router /api/redis/keys/{pattern} [get]
func (h *RedisHandler) getKeys(c *fiber.Ctx) error {
	pattern := c.Params("pattern")

	keys := h.redisServer.Keys(pattern)

	return c.JSON(api.GetKeysResponse{
		Keys: keys,
	})
}

// @Summary Set a hash field
// @Description Set a field in a hash stored at key
// @Tags redis
// @Accept json
// @Produce json
// @Param data body api.HSetKeyRequest true "Hash field data"
// @Success 200 {object} api.HSetKeyResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /api/redis/hset [post]
func (h *RedisHandler) hsetKey(c *fiber.Ctx) error {
	var req api.HSetKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	added := h.redisServer.HSet(req.Key, req.Field, req.Value)

	return c.JSON(api.HSetKeyResponse{
		Added: added == 1,
	})
}

// @Summary Get a hash field
// @Description Get the value of a field in a hash stored at key
// @Tags redis
// @Produce json
// @Param key path string true "Hash key"
// @Param field path string true "Hash field"
// @Success 200 {object} api.HGetKeyResponse
// @Failure 404 {object} api.ErrorResponse
// @Router /api/redis/hget/{key}/{field} [get]
func (h *RedisHandler) hgetKey(c *fiber.Ctx) error {
	key := c.Params("key")
	field := c.Params("field")

	value, ok := h.redisServer.HGet(key, field)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(api.ErrorResponse{
			Error: "Key or field not found",
		})
	}

	return c.JSON(api.HGetKeyResponse{
		Value: value,
	})
}

// @Summary Delete hash fields
// @Description Delete one or more fields from a hash stored at key
// @Tags redis
// @Accept json
// @Produce json
// @Param data body api.HDelKeyRequest true "Hash fields to delete"
// @Success 200 {object} api.HDelKeyResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /api/redis/hdel [post]
func (h *RedisHandler) hdelKey(c *fiber.Ctx) error {
	var req api.HDelKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: "Invalid request: " + err.Error(),
		})
	}

	count := h.redisServer.HDel(req.Key, req.Fields)

	return c.JSON(api.HDelKeyResponse{
		Count: count,
	})
}

// @Summary Get hash keys
// @Description Get all field names in a hash stored at key
// @Tags redis
// @Produce json
// @Param key path string true "Hash key"
// @Success 200 {object} api.HKeysResponse
// @Router /api/redis/hkeys/{key} [get]
func (h *RedisHandler) hkeysKey(c *fiber.Ctx) error {
	key := c.Params("key")

	fields := h.redisServer.HKeys(key)

	return c.JSON(api.HKeysResponse{
		Fields: fields,
	})
}

// @Summary Get hash length
// @Description Get the number of fields in a hash stored at key
// @Tags redis
// @Produce json
// @Param key path string true "Hash key"
// @Success 200 {object} api.HLenResponse
// @Router /api/redis/hlen/{key} [get]
func (h *RedisHandler) hlenKey(c *fiber.Ctx) error {
	key := c.Params("key")

	length := h.redisServer.HLen(key)

	return c.JSON(api.HLenResponse{
		Length: length,
	})
}

// @Summary Increment a key
// @Description Increment the integer value of a key by one
// @Tags redis
// @Produce json
// @Param key path string true "Key to increment"
// @Success 200 {object} api.IncrKeyResponse
// @Failure 400 {object} api.ErrorResponse
// @Router /api/redis/incr/{key} [post]
func (h *RedisHandler) incrKey(c *fiber.Ctx) error {
	key := c.Params("key")

	value, err := h.redisServer.Incr(key)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(api.IncrKeyResponse{
		Value: value,
	})
}
