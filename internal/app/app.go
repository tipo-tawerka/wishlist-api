package app

import (
	"github.com/go-playground/validator/v10"
	"github.com/tipo-tawerka/wishlist-api/config"
	httpController "github.com/tipo-tawerka/wishlist-api/internal/controller/http"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/handler"
	"github.com/tipo-tawerka/wishlist-api/internal/controller/http/middleware"
	authrepo "github.com/tipo-tawerka/wishlist-api/internal/repository/postgres/auth"
	itemsrepo "github.com/tipo-tawerka/wishlist-api/internal/repository/postgres/items"
	wishlistrepo "github.com/tipo-tawerka/wishlist-api/internal/repository/postgres/wishlist"
	authUsecase "github.com/tipo-tawerka/wishlist-api/internal/usecase/auth"
	itemsUsecase "github.com/tipo-tawerka/wishlist-api/internal/usecase/items"
	publicUsecase "github.com/tipo-tawerka/wishlist-api/internal/usecase/public"
	wishlistUsecase "github.com/tipo-tawerka/wishlist-api/internal/usecase/wishlist"
	"github.com/tipo-tawerka/wishlist-api/migrations"
	"github.com/tipo-tawerka/wishlist-api/pkg/auth"

	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	pgMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
)

func Run(cfg *config.Config, logger *zerolog.Logger) error {
	pool, err := pgxpool.New(context.Background(), cfg.Postgres.DSN())
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping postgres: %w", err)
	}

	if err := runMigrations(pool, logger); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	tokenManager, err := auth.NewTokenManager(cfg.Auth.JWTSecret)
	if err != nil {
		return fmt.Errorf("failed to create token manager: %w", err)
	}
	passwordHasher, err := auth.NewPasswordHasher(cfg.Auth.Pepper)
	if err != nil {
		return fmt.Errorf("failed to create password hasher: %w", err)
	}

	authRepository := authrepo.NewUserRepository(pool)
	itemsRepository := itemsrepo.NewItemsRepository(pool)
	wishlistRepository := wishlistrepo.NewWishlistRepository(pool)

	authUseCase := authUsecase.NewUseCase(authUsecase.Options{
		UserRepo:       authRepository,
		TokenManager:   tokenManager,
		PasswordHasher: passwordHasher,
		Logger:         logger,
	})
	itemsUseCase := itemsUsecase.NewUseCase(itemsRepository, wishlistRepository, logger)
	wishlistUseCase := wishlistUsecase.NewUseCase(wishlistRepository, logger)
	publicUseCase := publicUsecase.NewUseCase(wishlistRepository, itemsRepository, logger)

	h := handler.NewHandler(handler.Options{
		Auth:      authUseCase,
		Items:     itemsUseCase,
		WishList:  wishlistUseCase,
		Public:    publicUseCase,
		Validator: validator.New(),
	})

	authMiddleware := middleware.NewAuthMiddleware(tokenManager)

	router := httpController.NewRouter(h, authMiddleware)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Info().Str("port", cfg.HTTP.Port).Msg("server started")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("server failed")
		}
	}()

	<-quit
	logger.Info().Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(ctx)
}

func runMigrations(pool *pgxpool.Pool, logger *zerolog.Logger) error {
	srcDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("failed to create iofs source: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)

	dbDriver, err := pgMigrate.WithInstance(db, &pgMigrate.Config{})
	if err != nil {
		return fmt.Errorf("failed to create db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", srcDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info().Msg("migrations applied successfully")
	return nil
}
