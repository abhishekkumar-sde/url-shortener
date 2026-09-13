package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"url-shortener/encoder/bl"
	"url-shortener/encoder/dl"
	"url-shortener/encoder/endpoint"
	"url-shortener/encoder/inithandler"
	"url-shortener/encoder/svcparam/envvar"
	httptransport "url-shortener/encoder/transport/http"
	"url-shortener/pkg/ratelimiter"
	"url-shortener/pkg/redis"

	"github.com/spf13/viper"
)

func main() {

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
	viper.AutomaticEnv()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := inithandler.GetDynamoDb(ctx)
	if err != nil {
		logger.Error("dynamodb initialization failed", "error", err)
		os.Exit(1)
	}

	if err := db.EnsureTable(ctx); err != nil {
		logger.Error("dynamodb table initialization failed", "error", err)
		os.Exit(1)
	}

	redisClient, err := redis.NewRedisClient(ctx, viper.GetString(envvar.RedisHost), viper.GetString(envvar.RedisPort))
	if err != nil {
		logger.Error("redis initialization failed", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()

	repo := dl.NewURLRepository(db.Client, viper.GetString(envvar.DynamoDBTable))
	cache := redis.NewCache(redisClient, 24*time.Hour)
	limiter := ratelimiter.NewRateLimiter(redisClient)

	encoderBL := bl.NewEncoderBL(repo, cache, cache, viper.GetString(envvar.BaseURL))
	endpoint := endpoint.NewURLEndpoint(encoderBL, limiter)

	router := httptransport.NewRouter(endpoint)

	server := &http.Server{
		Addr:              ":" + viper.GetString(envvar.HTTPPort),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server started", "port", viper.GetString(envvar.HTTPPort))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}

	logger.Info("server stopped")
}
