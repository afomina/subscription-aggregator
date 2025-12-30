package main

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"subscription-aggregator/docs"
	"subscription-aggregator/internal/config"
	"subscription-aggregator/internal/db"
	"subscription-aggregator/internal/handler"
	"subscription-aggregator/internal/repository"
	"subscription-aggregator/internal/service"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Subscription Aggregator API
// @version 1.0
// @description REST API for managing user subscriptions.

// @host localhost:8080
// @BasePath /

func main() {
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	logger := logrus.New()
	level, _ := logrus.ParseLevel(cfg.LogLevel)
	logger.SetLevel(level)

	// Подключение к БД
	dbConn, err := db.NewPostgresDB(cfg)
	if err != nil {
		logger.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbConn.Close()

	// Запуск миграций
	migrationPath := os.Getenv("MIGRATIONS_PATH")
	if migrationPath == "" {
		migrationPath = "file://migrations" // путь по умолчанию
	}
	if err := runMigrations(cfg, migrationPath, logger); err != nil {
		logger.Fatalf("Migrations failed: %v", err)
	}

	// Инициализация сервиса
	repo := repository.NewSubscriptionRepo(dbConn)
	svc := service.NewSubscriptionService(repo)
	hdl := handler.NewHandler(svc, logger)

	// Настройка роутера
	r := gin.Default()

	docs.SwaggerInfo.Host = "localhost:" + strconv.Itoa(cfg.ServerPort)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	sub := r.Group("/subscriptions")
	{
		sub.POST("", hdl.CreateSubscription)
		sub.GET("", hdl.ListSubscriptions)
		sub.GET("/:id", hdl.GetSubscription)
		sub.PUT("/:id", hdl.UpdateSubscription)
		sub.DELETE("/:id", hdl.DeleteSubscription)
		sub.GET("/total-cost", hdl.GetTotalCost)
	}

	logger.Infof("Starting server on port %d", cfg.ServerPort)
	if err := r.Run(":" + strconv.Itoa(cfg.ServerPort)); err != nil {
		logger.Fatalf("Server failed: %v", err)
	}
}

func runMigrations(cfg *config.Config, migrationsPath string, logger *logrus.Logger) error {
	// Формируем DSN для migrate (без database/sql)
	query := url.Values{}
	query.Set("sslmode", "disable")
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
		query.Encode(),
	)

	logger.Infof("Applying migrations from %s", migrationsPath)
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Применяем все миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up failed: %w", err)
	}

	if err == migrate.ErrNoChange {
		logger.Info("Database is up to date")
	} else {
		logger.Info("Migrations applied successfully")
	}
	return nil
}