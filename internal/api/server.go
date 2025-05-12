package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/franky69420/crypto-oracle/internal/memory"
	"github.com/franky69420/crypto-oracle/pkg/utils/config"
	"github.com/franky69420/crypto-oracle/pkg/utils/logger"
)

// Server gère le serveur HTTP pour l'API
type Server struct {
	config      *config.APIConfig
	router      *mux.Router
	httpServer  *http.Server
	logger      *logger.Logger
	trustNetwork memory.MemoryOfTrust
}

// NewServer crée un nouveau serveur API
func NewServer(config *config.APIConfig, trustNetwork memory.MemoryOfTrust, logger *logger.Logger) *Server {
	router := mux.NewRouter()
	
	server := &Server{
		config:      config,
		router:      router,
		logger:      logger,
		trustNetwork: trustNetwork,
	}
	
	// Initialiser les routes
	server.initializeRoutes()
	
	return server
}

// initializeRoutes configure toutes les routes de l'API
func (s *Server) initializeRoutes() {
	// Configurer CORS
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	
	// Routes de base
	s.router.HandleFunc("/api/health", s.HealthCheck).Methods("GET")
	
	// Enregistrer les gestionnaires d'API
	activeWalletHandler := NewActiveWalletHandler(s.trustNetwork, s.logger)
	activeWalletHandler.RegisterRoutes(s.router)
	
	// Appliquer le middleware CORS
	s.router.Use(corsMiddleware.Handler)
	s.router.Use(s.loggingMiddleware)
}

// HealthCheck est un endpoint pour vérifier l'état du serveur
func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

// loggingMiddleware enregistre les informations sur les requêtes HTTP
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Appeler le gestionnaire suivant
		next.ServeHTTP(w, r)
		
		// Enregistrer après la requête
		s.logger.Info("HTTP Request",
			map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
				"duration_ms": time.Since(start).Milliseconds(),
			},
		)
	})
}

// Start démarre le serveur HTTP
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	
	s.httpServer = &http.Server{
		Addr:           addr,
		Handler:        s.router,
		ReadTimeout:    time.Duration(s.config.ReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(s.config.WriteTimeout) * time.Second,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}
	
	s.logger.Info("Démarrage du serveur API", map[string]interface{}{
		"address": addr,
	})
	
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	
	return nil
}

// Shutdown arrête proprement le serveur HTTP
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Arrêt du serveur API")
	
	// Créer un contexte avec un délai
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	
	// Arrêter le serveur
	return s.httpServer.Shutdown(shutdownCtx)
} 