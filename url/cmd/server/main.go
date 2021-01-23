package main

import (
	"context"
	"flag"
	"fmt"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/go-ozzo/ozzo-routing/v2/content"
	"github.com/go-ozzo/ozzo-routing/v2/cors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"url/internal/analytics"
	"url/internal/auth"
	"url/internal/config"
	"url/internal/errors"
	"url/internal/healthcheck"
	"url/internal/store"
	"url/internal/urlShortner"
	"url/pkg/accesslog"
	"url/pkg/jwt"
	"url/pkg/log"
	"url/pkg/redis"
)

// Version indicates the current version of the application.
var Version = "1.0.0"
var flagConfig = flag.String("config", "./config/local.yml", "path to the config file")

func main() {
	flag.Parse()
	// create root logger tagged with server version
	logger := log.New().With(nil, "version", Version)
	var err error
	// load application configurations
	config.Cfg, err = config.Load(*flagConfig, logger)
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	// redis client
	redisService, err := redis.New(
		config.Cfg.Redis.Host,
		config.Cfg.Redis.Port,
		config.Cfg.Redis.Password,
	)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}
	defer redisService.Pool.Close()

	// store database
	psqlStore, err := store.NewPostgresStore(store.PostgresConfig{
		Logger:   logger,
		Host:     config.Cfg.Postgres.Host,
		Port:     config.Cfg.Postgres.Port,
		User:     config.Cfg.Postgres.User,
		Password: config.Cfg.Postgres.Password,
		DBName:   config.Cfg.Postgres.DBName,
	})
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	jwtService, err := jwt.New(jwt.Options{
		AccessSecret:  config.Cfg.JwtRSAKeys.Access,
		RefreshSecret: config.Cfg.JwtRSAKeys.Refresh,
		Redis:         redisService,
	})
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}

	// build HTTP server
	bindAddress := fmt.Sprintf(":%v", config.Cfg.ServerPort)

	// create a new server
	s := http.Server{
		Addr:         bindAddress,                                                           // configure the bind address
		Handler:      buildHandler(logger, psqlStore, redisService, jwtService, config.Cfg), // set the default handler
		ReadTimeout:  5 * time.Second,                                                       // max time to read request from the client
		WriteTimeout: 10 * time.Second,                                                      // max time to write response to the client
		IdleTimeout:  120 * time.Second,                                                     // max time for connections using TCP Keep-Alive
	}

	// start the server
	go func() {
		logger.Infof("Starting server on port: %s", bindAddress)

		err := s.ListenAndServe()
		if err != nil {
			logger.Errorf("Error starting server: %s", err)
			os.Exit(1)
		}
	}()

	// trap sigterm or interrupt and gracefully shutdown the server
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	sig := <-done
	logger.Infof("Got signal: %s", sig)

	// gracefully shutdown the server, waiting max 30 seconds for current operations to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		logger.Errorf("Server Shutdown with Error: %s", err)
	} else {
		logger.Info("Server Shutdown gracefully")
	}
}

// buildHandler sets up the HTTP routing and builds an HTTP handler.
func buildHandler(logger log.Logger, psqlStore *store.PostgresStore, redisService *redis.Redis, jwtService *jwt.Auth, cfg *config.Config) http.Handler {
	router := routing.New()

	router.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		content.TypeNegotiator(content.JSON),
		cors.Handler(cors.AllowAll),
	)

	healthcheck.RegisterHandlers(router, Version)

	rg := router.Group("")

	authHandler := jwtService.Handler()

	auth.RegisterHandlers(
		rg.Group("/api/v1/users"),
		auth.NewService(psqlStore, auth.NewRepository(redisService, logger), logger, jwtService),
		logger,
	)

	analytics.RegisterHandlers(
		rg.Group("/api/v1/anal"),
		analytics.NewService(psqlStore, logger),
		logger, authHandler,
	)

	urlShortner.RegisterHandlers(
		rg.Group("/"),
		urlShortner.NewService(psqlStore, psqlStore, urlShortner.NewRepository(redisService, logger), logger),
		logger, authHandler,
	)
	return router
}
