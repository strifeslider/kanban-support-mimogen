package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/user/kanban-saas/pkg/auth"
	"github.com/user/kanban-saas/pkg/database"
	appmiddleware "github.com/user/kanban-saas/pkg/middleware"
	"github.com/user/kanban-saas/services/support/internal/handler"
	"github.com/user/kanban-saas/services/support/internal/repository"
	"github.com/user/kanban-saas/services/support/internal/service"
)

func main() {
	env := getEnv("ENV", "local")
	port := getEnv("PORT", "8084")

	logger := setupLogger(env)
	logger.Info("starting support service", "env", env)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.NewPostgresPool(ctx, database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     5432,
		User:     getEnv("DB_USER", "kanban"),
		Password: getEnv("DB_PASSWORD", "kanban_dev_password"),
		Database: getEnv("DB_NAME", "kanban_support"),
		MaxConns: 10,
		MinConns: 2,
	})
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	runMigrations(ctx, db, logger)

	jwtCfg := auth.JWTConfig{
		Secret: getEnv("JWT_SECRET", "dev-secret-key-change-in-production"),
	}

	ticketRepo := repository.NewTicketRepository(db)
	agentRepo := repository.NewAgentRepository(db)

	supportService := service.NewSupportService(ticketRepo, agentRepo)

	ticketHandler := handler.NewTicketHandler(supportService)
	agentHandler := handler.NewAgentHandler(supportService)
	internalHandler := handler.NewInternalEventHandler(supportService)

	r := chi.NewRouter()

	allowedOrigins := appmiddleware.ParseOrigins(getEnv("CORS_ORIGINS", "http://localhost:3000"))
	r.Use(appmiddleware.CORS(allowedOrigins))
	r.Use(appmiddleware.Logging(logger))
	r.Use(appmiddleware.Recovery(logger))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	handler.SetupRoutes(r, ticketHandler, agentHandler, internalHandler, jwtCfg)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("support service listening", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down support service...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}
	logger.Info("support service stopped")
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func setupLogger(env string) *slog.Logger {
	var handler slog.Handler
	switch env {
	case "local":
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	case "dev":
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	default:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	return slog.New(handler)
}

func runMigrations(ctx context.Context, db *pgxpool.Pool, logger *slog.Logger) {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS tickets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			workspace_id UUID NOT NULL,
			subject VARCHAR(500) NOT NULL,
			status VARCHAR(50) NOT NULL DEFAULT 'open',
			priority VARCHAR(50) NOT NULL DEFAULT 'medium',
			source VARCHAR(50) NOT NULL,
			channel_id VARCHAR(255),
			user_id UUID NOT NULL,
			agent_id UUID,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			resolved_at TIMESTAMPTZ,
			deleted_at TIMESTAMPTZ
		);`,
		`CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status) WHERE deleted_at IS NULL;`,
		`CREATE TABLE IF NOT EXISTS ticket_messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
			sender_id UUID NOT NULL,
			sender_type VARCHAR(20) NOT NULL,
			content TEXT NOT NULL,
			platform VARCHAR(50),
			external_id VARCHAR(255),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS ticket_status_log (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
			old_status VARCHAR(50),
			new_status VARCHAR(50) NOT NULL,
			changed_by UUID NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS support_agents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL UNIQUE,
			max_tickets INTEGER NOT NULL DEFAULT 10,
			is_online BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(ctx, m); err != nil {
			logger.Error("migration failed", "error", err)
			os.Exit(1)
		}
	}
	logger.Info("migrations completed")
}
