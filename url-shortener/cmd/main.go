package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"url-shortener/internal/config"
	httprouter "url-shortener/internal/http/router"
	"url-shortener/internal/logger"
	"url-shortener/internal/service"
	"url-shortener/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = log.Sync()
	}()

	ctx := context.Background()
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("create database pool failed", zap.Error(err))
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatal("connect database failed", zap.Error(err))
	}

	linkStorage := storage.NewPostgresLinkStorage(db, log)
	userStorage := storage.NewPostgresUserStorage(db, log)
	linkService := service.NewLinkService(linkStorage, cfg.BaseURL, log)
	authService := service.NewAuthService(userStorage, cfg.JWTSecret, cfg.JWTAccessTTL, log)
	router := httprouter.New(linkService, authService, log)

	log.Info("starting server", zap.String("addr", cfg.HTTPAddr))
	if err := router.Run(cfg.HTTPAddr); err != nil {
		log.Fatal("server stopped", zap.Error(err))
	}
}
