package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"go-pc-rem/internal/config"
)

// NewRouter creates and configures the HTTP router
func NewRouter(cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Middleware
	// r.Use(middleware.Logger) // Disabled for better performance on Windows
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// CORS - allow all for now
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Create handlers with config
	statsHandler := NewStatsHandler(cfg)
	powerHandler := NewPowerHandler(cfg)

	// Routes
	r.Route("/rog", func(r chi.Router) {
		// Health check
		r.Get("/status", StatusHandler)

		// Stats endpoints
		if cfg.Features.EnableStats {
			r.Get("/stats", statsHandler.GetFullStats)
			r.Get("/stats/memory", statsHandler.GetMemoryStats)
			r.Get("/stats/cpu", statsHandler.GetCpuStats)
			r.Get("/stats/disk", statsHandler.GetDiskStats)
		}

		// Power endpoints
		r.Route("/pw", func(r chi.Router) {
			if cfg.Features.EnableShutdown {
				r.Post("/shutdown", powerHandler.Shutdown)
			}
			if cfg.Features.EnableRestart {
				r.Post("/restart", powerHandler.Restart)
			}
			if cfg.Features.EnableHibernate {
				r.Post("/hb", powerHandler.Hibernate)
			}
		})
	})

	return r
}
