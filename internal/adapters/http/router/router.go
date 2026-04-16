package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"crud-api/internal/adapters/http/middleware"
	userhandlers "crud-api/internal/adapters/http/user"
)

func New(userHandlers *userhandlers.Handlers, db *gorm.DB, rdb *redis.Client, env string) *gin.Engine {
	r := gin.New()

	r.Use(middleware.RequestID())
	r.Use(middleware.SlogLogger())
	r.Use(middleware.ErrorHandler())
	r.Use(gin.Recovery())

	// CORS config
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "X-Request-ID")
	r.Use(cors.New(corsConfig))

	r.GET("/healthz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		status := http.StatusOK
		checks := gin.H{"status": "ok"}

		// Check DB
		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			status = http.StatusServiceUnavailable
			checks["database"] = "down"
		} else {
			checks["database"] = "up"
		}

		// Check Redis (if enabled)
		if rdb != nil {
			if rdb.Ping(ctx).Err() != nil {
				status = http.StatusServiceUnavailable
				checks["redis"] = "down"
			} else {
				checks["redis"] = "up"
			}
		}

		c.JSON(status, checks)
	})

	v1 := r.Group("/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("", userHandlers.Create)
			users.GET("", userHandlers.List)
			users.GET("/:id", userHandlers.Get)
			users.PUT("/:id", userHandlers.Update)
			users.DELETE("/:id", userHandlers.Delete)
		}
	}

	return r
}
