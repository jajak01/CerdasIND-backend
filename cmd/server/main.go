package main

import (
	"log"
	"os"
	"time"

	"cerdasind-backend/internal/handler"
	"cerdasind-backend/internal/middleware"
	"cerdasind-backend/internal/model"
	"cerdasind-backend/internal/repository"
	"cerdasind-backend/internal/service"
	"cerdasind-backend/pkg/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "cerdasind-backend/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           CerdasIND API
// @version         1.0
// @description     Backend API for CerdasIND Examination System
// @BasePath        /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	db := database.InitDB()
	defer db.Close()

	// Repositories
	userRepo := repository.NewUserRepository(db)
	jenjangRepo := repository.NewJenjangRepository(db)
	mapelRepo := repository.NewMapelRepository(db)
	bundleRepo := repository.NewBundleRepository(db)
	soalRepo := repository.NewSoalRepository(db)
	historyRepo := repository.NewHistoryRepository(db)

	// Services
	authService := service.NewAuthService(userRepo)
	participantService := service.NewParticipantService(jenjangRepo, mapelRepo, bundleRepo, soalRepo, historyRepo)
	adminService := service.NewAdminService(db, bundleRepo, soalRepo, historyRepo, userRepo, jenjangRepo, mapelRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	participantHandler := handler.NewParticipantHandler(participantService)
	adminHandler := handler.NewAdminHandler(adminService)

	// Router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// Public Auth
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/register", authHandler.Register)
		}

		// Participant Area (JWT Protected)
		participant := api.Group("/")
		participant.Use(middleware.AuthMiddleware())
		{
			participant.GET("/jenjang", participantHandler.GetJenjang)
			participant.GET("/jenjang/:id/mapel", participantHandler.GetMapel)
			participant.GET("/mapel/:id/bundles", participantHandler.GetBundles)
			participant.GET("/bundles/:id/soal", participantHandler.GetSoal)
			participant.POST("/bundles/:id/submit", participantHandler.Submit)
			participant.GET("/users/history", participantHandler.GetHistory)
			participant.GET("/bundles/:id/review", participantHandler.GetReview)
		}

		// Admin Area (JWT Protected + Role Admin)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.RoleMiddleware(model.RoleAdmin))
		{
			admin.GET("/bundles", adminHandler.GetBundles)
			admin.POST("/bundles/upload", adminHandler.UploadBundle)
			admin.GET("/bundles/:id/export", adminHandler.ExportBundle)
			admin.PUT("/bundles/:id/update", adminHandler.UpdateBundle)
			admin.GET("/submissions", adminHandler.GetSubmissions)
			admin.GET("/submissions/:history_id", adminHandler.GetSubmissionDetail)
			admin.PUT("/submissions/:history_id/grade", adminHandler.GradeSubmission)
		}
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
