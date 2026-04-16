package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	httprouter "crud-api/internal/adapters/http/router"
	httpuser "crud-api/internal/adapters/http/user"
	postgresadapter "crud-api/internal/adapters/postgres"
	redisadapter "crud-api/internal/adapters/redis"
	applicationuser "crud-api/internal/application/user"
	userdecorator "crud-api/internal/application/user/decorator"
	"crud-api/internal/config"
	"crud-api/internal/db"
	"crud-api/internal/ports"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	var handler slog.Handler
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	gormDB, err := gorm.Open(postgres.Open(cfg.DB.DSN), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		slog.Error("connect db", "error", err)
		os.Exit(1)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		slog.Error("db handle", "error", err)
		os.Exit(1)
	}
	sqlDB.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.DB.ConnMaxLifetime)

	if err := db.AutoMigrate(gormDB); err != nil {
		slog.Error("migrate", "error", err)
		os.Exit(1)
	}

	var rdb *redis.Client
	var userCache ports.UserCache
	if cfg.Redis.Addr != "" {
		rdb = redis.NewClient(&redis.Options{Addr: cfg.Redis.Addr})
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := rdb.Ping(ctx).Err(); err != nil {
			slog.Warn("redis disabled (ping failed)", "error", err)
		} else {
			userCache = redisadapter.NewUserCache(rdb)
		}
	}

	repo := postgresadapter.NewUserRepository(gormDB)
	var svc ports.UserService
	svc = applicationuser.NewService(repo)
	if userCache != nil {
		svc = userdecorator.NewCacheDecorator(svc, userCache, cfg.Redis.UserTTL)
	}
	handlers := httpuser.NewHandlers(svc)

	router := httprouter.New(handlers, gormDB, rdb, cfg.AppEnv)

	if cfg.AppEnv != "production" {
		pprof.Register(router, "/debug/pprof")
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr, "env", cfg.AppEnv)
		errCh <- srv.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("shutdown signal", "signal", sig.String())
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}

	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(stopCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	slog.Info("server stopped")
}
