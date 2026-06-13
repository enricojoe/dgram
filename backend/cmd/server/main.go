package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/enricojoe/dgram/backend/internal/config"
	"github.com/enricojoe/dgram/backend/internal/controller"
	"github.com/enricojoe/dgram/backend/internal/db"
	"github.com/enricojoe/dgram/backend/internal/middleware"
	"github.com/enricojoe/dgram/backend/internal/repository"
	"github.com/enricojoe/dgram/backend/internal/service"
)

func main() {
	cfg := config.Load()

	// Connect to Postgres and apply migrations before serving traffic.
	conn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err := db.Migrate(conn); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	// Repositories -> services -> controllers.
	userRepo := repository.NewUserRepository(conn)
	diagramRepo := repository.NewDiagramRepository(conn)

	schemas := service.NewSchemaService()
	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	diagramService := service.NewDiagramService(diagramRepo)

	authController := controller.NewAuthController(authService)
	diagramController := controller.NewDiagramController(diagramService)

	router := gin.Default()
	router.Use(middleware.CORS(cfg.CORSOrigin))

	api := router.Group("/api")

	// Public routes.
	controller.NewHealthController().Register(api)
	controller.NewParseController(schemas).Register(api)
	controller.NewGenerateController(schemas).Register(api)
	authController.RegisterPublic(api)
	diagramController.RegisterPublic(api)

	// Protected routes (require a valid access token).
	protected := api.Group("")
	protected.Use(middleware.RequireAuth(cfg.JWTSecret))
	authController.RegisterProtected(protected)
	diagramController.Register(protected)

	addr := ":" + cfg.Port
	log.Printf("DGram backend listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
