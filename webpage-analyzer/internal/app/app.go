package app

import (
	"fmt"
	"net/http"

	"webpage-analyzer/internal/config"
)

// App contains the runtime HTTP dependencies for the application.
type App struct {
	Handler http.Handler
	Address string
	Port    string
}

// New builds the application from runtime configuration.
func New(cfg *config.Config) (*App, error) {
	switch cfg.Service.Role {
	case "web":
		return newWebApp(cfg)
	case "analysis":
		return newAnalysisApp(cfg)
	default:
		return nil, fmt.Errorf("unsupported service role: %s", cfg.Service.Role)
	}
}

func newRuntimeApp(port string, handler http.Handler) *App {
	return &App{
		Handler: handler,
		Address: ":" + port,
		Port:    port,
	}
}
