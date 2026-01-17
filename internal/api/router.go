package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/azaek/cntrl/internal/config"
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
	mediaHandler := NewMediaHandler(cfg)

	// Routes
	r.Route("/api", func(r chi.Router) {
		// Health check
		r.Get("/status", StatusHandler) // StatusHandler handles GET /api/status

		// Processes
		r.Get("/processes", statsHandler.GetProcessList)

		// New System Stats
		r.Get("/system", statsHandler.GetSystemInfo)
		r.Get("/usage", statsHandler.GetSystemUsage)

		// Media
		r.Route("/media", func(r chi.Router) {
			r.Post("/control", mediaHandler.HandleControl)
			r.Get("/status", mediaHandler.HandleStatus)
		})

		// Stats endpoints
		r.Route("/stats", func(r chi.Router) {
			r.Get("/", statsHandler.GetFullStats)
			r.Get("/memory", statsHandler.GetMemoryStats)
			r.Get("/cpu", statsHandler.GetCpuStats)
			r.Get("/disk", statsHandler.GetDiskStats)
		})

		// Power endpoints
		r.Route("/pw", func(r chi.Router) {
			r.Post("/shutdown", powerHandler.Shutdown)
			r.Post("/restart", powerHandler.Restart)
			r.Post("/hibernate", powerHandler.Hibernate)
			r.Post("/sleep", powerHandler.Sleep)
		})
	})

	return r
}
