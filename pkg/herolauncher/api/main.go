// Package api contains API handlers for HeroLauncher
package api

// @title HeroLauncher API
// @version 1.0
// @description API for HeroLauncher - a modular service manager
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@freeflowuniverse.org
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:9001
// @BasePath /api
// @schemes http https

// This file exists solely to provide Swagger documentation
// and to ensure all API handlers are included in the documentation

// AdminHandler handles admin-related API routes
// @Router /api/hardware-stats [get]
// @Router /api/process-stats [get]

// ServiceHandler handles service-related API routes
// @Router /api/services/running [get]
// @Router /api/services/start [post]
// @Router /api/services/stop [post]
// @Router /api/services/restart [post]
// @Router /api/services/delete [post]
// @Router /api/services/logs [post]
// @Router /admin/services/ [get]
// @Router /admin/services/data [get]
// @Router /admin/services/running [get]
// @Router /admin/services/start [post]
// @Router /admin/services/stop [post]
// @Router /admin/services/restart [post]
// @Router /admin/services/delete [post]
// @Router /admin/services/logs [post]

// ExecutorHandler handles command execution API routes
// @Router /api/executor/execute [post]
// @Router /api/executor/jobs [get]
// @Router /api/executor/jobs/{id} [get]

// JetHandler handles Jet template API routes
// @Router /api/jet/validate [post]

// RedisHandler handles Redis API routes
// @Router /api/redis/set [post]
// @Router /api/redis/get/{key} [get]
// @Router /api/redis/del/{key} [delete]
// @Router /api/redis/keys/{pattern} [get]
// @Router /api/redis/hset [post]
// @Router /api/redis/hget/{key}/{field} [get]
// @Router /api/redis/hdel [post]
// @Router /api/redis/hkeys/{key} [get]
// @Router /api/redis/hgetall/{key} [get]
